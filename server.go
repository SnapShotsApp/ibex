/* Copyright 2016 Snapshots LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const re = `(?i)/?uploads/(?P<env>\w+)/(?P<username>\w+)?/?picture/attachment/(?P<id>\d+)/(?P<name>\w+)(?:/[a-zA-Z0-9]+)?$`
const watermarkPathFmt = "%s/uploads/%s/%s/%d/%s"
const photographerInfoPathPart = "photographer_info/picture"
const watermarkPathPart = "watermark/logo"
const picturePathFmt = "%s/uploads/%s/picture/attachment/%d/%s"

var pathMatcher *regexp.Regexp

type imagizerHandler struct {
	imagizerHost    *url.URL
	config          *Config
	db              *DB
	logger          ILogger
	statsChan       chan *stat
	responseTimeout time.Duration
}

func init() {
	pathMatcher = regexp.MustCompile(re)
}

// Start initializes and then starts the HTTP server
func Start(c *Config, logger ILogger, statsChan chan *stat) {
	db, err := NewDB(c)
	logger.HandleErr(err)

	imagizerHost, err := url.Parse(c.ImagizerHost)
	logger.HandleErr(err)

	handler := imagizerHandler{
		imagizerHost:    imagizerHost,
		config:          c,
		db:              db,
		logger:          logger,
		statsChan:       statsChan,
		responseTimeout: 10 * time.Second,
	}

	s := &http.Server{
		Addr:         c.BindAddr(),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: handler.responseTimeout,
	}

	logger.Info("Listening on %s", s.Addr)
	err = s.ListenAndServe()
	logger.HandleErr(err)
}

type errorResponse struct {
	err    error
	status int
}

func validateAndExtractPath(req *http.Request) (parts map[string]string, err error) {
	if req.Method != http.MethodGet || !pathMatcher.MatchString(req.URL.Path) {
		err = fmt.Errorf("Malformed Path: %s", req.URL.Path)
		return
	}

	parts = extractPathPartsToMap(req.URL.Path)

	username := parts["username"]
	isDev := (parts["env"] == "development")
	if isDev != (len(username) > 0) {
		err = fmt.Errorf("Malformed Path: %s", req.URL.Path)
	}

	return
}

func (h imagizerHandler) handleRequest(ctx context.Context, req *http.Request, w http.ResponseWriter, done chan string, errChan chan errorResponse) {
	innerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger := innerCtx.Value("logger").(ILogger)

	parts, err := validateAndExtractPath(req)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusNotFound}
		return
	}
	logger.Debug("URL Parts: %v", parts)

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		cancel()
		errChan <- errorResponse{fmt.Errorf("Version not found with name %s", parts["name"]), http.StatusNotFound}
		return
	}
	logger.Debug("Version found: %v", version)

	pictureID, err := strconv.Atoi(parts["id"])
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}

	_, err = h.db.loadPicture(innerCtx, pictureID)
	if err != nil {
		cancel()
		var status int

		switch err.(type) {
		case noRowsErr:
			status = http.StatusNotFound
		default:
			status = http.StatusInternalServerError
		}

		errChan <- errorResponse{err, status}
		return
	}

	proxy, err := h.imagizerURL(innerCtx, version, parts)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}
	imagizerReq, err := http.NewRequest("GET", proxy.String(), nil)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}
	imagizerReq = imagizerReq.WithContext(innerCtx)

	resp, err := http.DefaultClient.Do(imagizerReq)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}
	logger.Debug("Imagizer response: %v", resp)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}

	done <- parts["name"]
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.responseTimeout)
	ctx = context.WithValue(ctx, "logger", h.logger)
	defer cancel()
	req = req.WithContext(ctx)

	done := make(chan string)
	errChan := make(chan errorResponse)
	go h.handleRequest(ctx, req, w, done, errChan)

	handleTimeout := func() {
		http.Error(w, "timeout", http.StatusGatewayTimeout)
		h.statsChan <- &stat{StatTimeout, ""}
	}

	select {
	case name := <-done:
		h.statsChan <- &stat{StatServedPicture, name}
	case errResp := <-errChan:
		if errResp.err == context.DeadlineExceeded || errResp.err == context.Canceled {
			handleTimeout()
			return
		}

		http.Error(w, errResp.err.Error(), errResp.status)
		h.statsChan <- &stat{StatBadRequest, ""}
	case <-ctx.Done():
		handleTimeout()
	}
}

func (h imagizerHandler) imagizerURL(ctx context.Context, version map[string]interface{}, parts map[string]string) (url.URL, error) {
	vals := url.Values{}
	retURL := url.URL{}

	pictureID, err := strconv.Atoi(parts["id"])
	if err != nil {
		return retURL, err
	}

	for key, val := range version {
		if key == "watermark" && val == true {
			isPhotogImage, err := h.isPhotographerImage(ctx, pictureID)
			if err != nil {
				return retURL, err
			}

			if isPhotogImage {
				wm, err := h.getWatermarkInfo(ctx, pictureID, parts["env"])
				if err != nil {
					return retURL, err
				}

				vals.Add("mark", wm.logo.String)
				if wm.scale.Valid {
					vals.Add("mark_scale", strconv.FormatInt(wm.scale.Int64, 10))
				} else {

					vals.Add("mark_scale", "0")
				}
				if wm.offset.Valid {
					vals.Add("mark_offset", strconv.FormatInt(wm.offset.Int64, 10))
				} else {
					vals.Add("mark_offset", "0")
				}
				if wm.alpha.Valid {
					vals.Add("mark_alpha", strconv.FormatInt(wm.alpha.Int64, 10))
				} else {
					vals.Add("mark_alpha", "0")
				}
				vals.Add("mark_pos", wm.position.String)
			}

			continue
		}

		if key == "function_name" || key == "name" || key == "only_shrink_larger" {
			continue
		}

		switch val := val.(type) {
		default:
			return retURL, fmt.Errorf("Unexpected type %T for %v", val, val)
		case string:
			vals.Add(key, val)
		case int:
			vals.Add(key, strconv.Itoa(val))
		case float64:
			vals.Add(key, strconv.Itoa(int(val)))
		case bool:
			vals.Add(key, fmt.Sprintf("%t", val))
		}
	}

	path, err := h.pathForImage(ctx, parts)
	if err != nil {
		return retURL, err
	}

	retURL.Scheme = h.imagizerHost.Scheme
	retURL.Host = h.imagizerHost.Host
	retURL.Path = path
	retURL.RawQuery = vals.Encode()

	return retURL, nil
}

func (h imagizerHandler) isPhotographerImage(ctx context.Context, id int) (bool, error) {
	pic, err := h.db.loadPicture(ctx, id)
	if err != nil {
		return false, err
	}

	ev, err := h.db.loadEvent(ctx, pic.eventID)
	if err != nil {
		return false, err
	}

	is := pic.userID == ev.ownerID
	return is, nil
}

func (h imagizerHandler) getWatermarkInfo(ctx context.Context, id int, env string) (watermark, error) {
	pic, err := h.db.loadPicture(ctx, id)
	if err != nil {
		return watermark{}, err
	}
	pi, err := h.db.loadPhotographerInfo(ctx, pic.userID)
	if err != nil {
		return watermark{}, err
	}
	wm, err := h.db.loadWatermark(ctx, pi.id)

	if err != nil {
		wm = watermark{
			logo:     newNullString("https://www.snapshots.com/images/icon.png"),
			disabled: false,
			alpha:    newNullInt64(70),
			scale:    newNullInt64(15),
			offset:   newNullInt64(3),
			position: newNullString("bottom,right"),
		}

		if pi.picture.Valid {
			wm.logo = newNullString(fmt.Sprintf(watermarkPathFmt,
				h.config.CDNHost, env, photographerInfoPathPart, pi.id, pi.picture.String))
		}
	} else {
		if wm.logo.Valid {
			wm.logo = newNullString(fmt.Sprintf(watermarkPathFmt,
				h.config.CDNHost, env, watermarkPathPart, wm.id, wm.logo.String))
		}
	}

	return wm, nil
}

func (h imagizerHandler) pathForImage(ctx context.Context, parts map[string]string) (string, error) {
	pictureID, err := strconv.Atoi(parts["id"])
	if err != nil {
		return "", err
	}

	pic, err := h.db.loadPicture(ctx, pictureID)
	if err != nil {
		return "", err
	}

	env := parts["env"]

	var envAndUsername string

	if env == "development" {
		envAndUsername = fmt.Sprintf("%s/%s", env, parts["username"])
	} else {
		envAndUsername = env
	}

	path := fmt.Sprintf(picturePathFmt,
		BucketNames[env], envAndUsername, pictureID, pic.attachment)
	return path, nil
}

func extractPathPartsToMap(path string) map[string]string {
	names := pathMatcher.SubexpNames()[1:]
	matches := pathMatcher.FindStringSubmatch(path)

	md := map[string]string{}

	for i, s := range matches[1:] {
		md[names[i]] = s
	}

	return md
}
