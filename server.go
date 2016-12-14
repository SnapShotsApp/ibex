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

	"github.com/satori/go.uuid"
)

const re = `(?i)/?uploads/(?P<env>\w+)/(?P<username>\w+)?/?picture/attachment/(?P<id>\d+)/(?P<name>\w+)(?:/[a-zA-Z0-9_-]+)?$`
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
		responseTimeout: 20 * time.Second,
	}

	s := &http.Server{
		Addr:    c.BindAddr(),
		Handler: handler,
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

type requestInfo struct {
	pictureID   int
	env         string
	username    string
	versionName string
	versionInfo map[string]interface{}
	info        pictureInfo
}

func (r requestInfo) isPhotographerImage() bool {
	return r.info.userID == r.info.ownerID
}

func (h imagizerHandler) handleRequest(ctx context.Context, req *http.Request, w http.ResponseWriter, done chan string, errChan chan errorResponse) {
	innerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger := innerCtx.Value("logger").(ILogger)
	logger.Info("START [GET] %s", req.URL.Path)

	parts, err := validateAndExtractPath(req)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusNotFound}
		return
	}
	logger.Debug("URL Parts: %+v", parts)

	rinfo := requestInfo{
		env:         parts["env"],
		username:    parts["username"],
		versionName: parts["name"],
	}

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		cancel()
		errChan <- errorResponse{fmt.Errorf("Version not found with name %s", parts["name"]), http.StatusNotFound}
		return
	}
	logger.Debug("Version found: %+v", version)

	rinfo.versionInfo = version

	pictureID, err := strconv.Atoi(parts["id"])
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}
	rinfo.pictureID = pictureID

	info, err := h.db.loadPictureInfo(ctx, rinfo.pictureID)
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
	rinfo.info = info

	proxy, err := h.imagizerURL(innerCtx, rinfo)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}
	logger.Debug("Imagizer URL: %+v", proxy)

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
	logger.Debug("Imagizer response: %+v", resp)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		cancel()
		errChan <- errorResponse{err, http.StatusInternalServerError}
		return
	}

	started := innerCtx.Value("startTime").(time.Time)
	logger.Info(fmt.Sprintf("FINISH [GET] %s (%s)", req.URL.Path, time.Since(started)))
	done <- parts["name"]
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.responseTimeout)
	innerLogger := h.logger.Sub()
	innerLogger.SetPrefix(fmt.Sprintf("[%s]", uuid.NewV4().String()))
	ctx = context.WithValue(ctx, "logger", innerLogger)
	ctx = context.WithValue(ctx, "startTime", time.Now())
	defer cancel()
	req = req.WithContext(ctx)

	done := make(chan string)
	errChan := make(chan errorResponse)
	go h.handleRequest(ctx, req, w, done, errChan)

	handleTimeout := func(err string) {
		innerLogger.Warn("timeout: %s", err)
		http.Error(w, "timeout", http.StatusGatewayTimeout)
		h.statsChan <- &stat{StatTimeout, ""}
	}

	select {
	case name := <-done:
		cancel()
		h.statsChan <- &stat{StatServedPicture, name}
	case errResp := <-errChan:
		if errResp.err == context.DeadlineExceeded || errResp.err == context.Canceled {
			handleTimeout(errResp.err.Error())
			return
		}

		http.Error(w, errResp.err.Error(), errResp.status)
		h.statsChan <- &stat{StatBadRequest, ""}
	case <-ctx.Done():
		handleTimeout(ctx.Err().Error())
	}
}

func (h imagizerHandler) imagizerURL(ctx context.Context, rinfo requestInfo) (url.URL, error) {
	vals := url.Values{}
	retURL := url.URL{}

	for key, val := range rinfo.versionInfo {
		if key == "watermark" && val == true {
			if rinfo.isPhotographerImage() {
				wm := h.getCanonicalWatermark(rinfo)

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
			return retURL, fmt.Errorf("Unexpected type %T for %+v", val, val)
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

	path, err := h.pathForImage(ctx, rinfo)
	if err != nil {
		return retURL, err
	}

	retURL.Scheme = h.imagizerHost.Scheme
	retURL.Host = h.imagizerHost.Host
	retURL.Path = path
	retURL.RawQuery = vals.Encode()

	return retURL, nil
}

func (h imagizerHandler) getCanonicalWatermark(rinfo requestInfo) watermark {
	wm := rinfo.info.mark

	if !wm.logo.Valid {
		wm = watermark{
			logo:     newNullString("https://www.snapshots.com/images/icon.png"),
			disabled: newNullBool(false),
			alpha:    newNullInt64(70),
			scale:    newNullInt64(15),
			offset:   newNullInt64(3),
			position: newNullString("bottom,right"),
		}

		if rinfo.info.oldMark.Valid {
			wm.logo = newNullString(fmt.Sprintf(watermarkPathFmt, h.config.CDNHost,
				rinfo.env, photographerInfoPathPart, rinfo.info.photographerInfoID.Int64,
				rinfo.info.oldMark.String))
		}
	} else {
		if wm.logo.Valid {
			wm.logo = newNullString(fmt.Sprintf(watermarkPathFmt,
				h.config.CDNHost, rinfo.env, watermarkPathPart,
				wm.id.Int64, wm.logo.String))
		}
	}

	return wm
}

func (h imagizerHandler) pathForImage(ctx context.Context, rinfo requestInfo) (string, error) {
	var envAndUsername string

	if rinfo.env == "development" {
		envAndUsername = fmt.Sprintf("%s/%s", rinfo.env, rinfo.username)
	} else {
		envAndUsername = rinfo.env
	}

	path := fmt.Sprintf(picturePathFmt,
		BucketNames[rinfo.env], envAndUsername, rinfo.pictureID, rinfo.info.attachment)
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
