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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const re = `(?i)/?uploads/(?P<env>\w+)/(?P<username>\w+)?/?picture/attachment/(?P<id>\d+)/(?P<name>\w+)$`
const watermarkPathFmt = "%s/uploads/%s/%s/%d/%s"
const photographerInfoPathPart = "photographer_info/picture"
const watermarkPathPart = "watermark/logo"
const picturePathFmt = "%s/uploads/%s/picture/attachment/%d/%s"

var pathMatcher *regexp.Regexp

type imagizerHandler struct {
	imagizerHost *url.URL
	config       *Config
	db           *DB
	logger       ILogger
}

func init() {
	pathMatcher = regexp.MustCompile(re)
}

// Start initializes and then starts the HTTP server
func Start(c *Config, logger ILogger) {
	db, err := NewDB(c)
	logger.HandleErr(err)

	imagizerHost, err := url.Parse(c.ImagizerHost)
	logger.HandleErr(err)

	handler := imagizerHandler{
		imagizerHost: imagizerHost,
		config:       c,
		db:           db,
		logger:       logger,
	}

	s := &http.Server{
		Addr:    c.BindAddr(),
		Handler: handler,
	}

	logger.Info("Listening on %s", s.Addr)
	err = s.ListenAndServe()
	logger.HandleErr(err)
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	notFound := func(msg string) {
		h.logger.Warn(msg)
		http.NotFound(w, req)
	}

	if req.Method != http.MethodGet {
		notFound("Only GET requests allowed")
		return
	}

	if !pathMatcher.MatchString(req.URL.Path) {
		notFound(fmt.Sprintf("Malformed Path: %s", req.URL.Path))
		return
	}

	parts := extractPathPartsToMap(req.URL.Path)
	h.logger.Debug("URL Parts: %v", parts)

	username, _ := parts["username"]
	isDev := (parts["env"] == "development")
	if isDev != (len(username) > 0) {
		notFound(fmt.Sprintf("Malformed Path: %s", req.URL.Path))
		return
	}

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		notFound(fmt.Sprintf("Version not found with name %s", parts["name"]))
		return
	}
	h.logger.Debug("Version found: %v", version)

	pictureID, err := strconv.Atoi(parts["id"])
	h.logger.HandleErr(err)

	pic := h.db.loadPicture(pictureID, h.logger)
	if pic.eventID == 0 {
		h.logger.Warn("Picture not found for ID %d", pictureID)
		http.NotFound(w, req)
		return
	}

	proxy := h.imagizerURL(version, parts)
	resp, err := http.Get(proxy.String())
	defer h.logger.CloseQuietly(resp.Body)
	h.logger.HandleErr(err)

	h.logger.Debug("Imagizer response: %v", resp)

	i, err := io.Copy(w, resp.Body)
	h.logger.HandleErr(err)
	h.logger.Debug("Copied %d bytes to output from Imagizer response", i)
}

func (h imagizerHandler) imagizerURL(version map[string]interface{}, parts map[string]string) url.URL {
	vals := url.Values{}
	pictureID, err := strconv.Atoi(parts["id"])
	h.logger.HandleErr(err)

	for key, val := range version {
		if key == "watermark" {
			if val == true && h.isPhotographerImage(pictureID) {
				wm := h.getWatermarkInfo(pictureID, parts["env"])
				vals.Add("mark", wm.logo.String)
				vals.Add("mark_scale", strconv.Itoa(wm.scale))
				vals.Add("mark_pos", wm.position)
				vals.Add("mark_offset", strconv.Itoa(wm.offset))
				vals.Add("mark_alpha", strconv.Itoa(wm.alpha))
			}

			continue
		}

		if key == "function_name" || key == "name" || key == "only_shrink_larger" {
			continue
		}

		switch val := val.(type) {
		default:
			h.logger.Fatal("Unexpected type %T for %v", val, val)
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

	retURL := url.URL{}
	retURL.Scheme = h.imagizerHost.Scheme
	retURL.Host = h.imagizerHost.Host
	retURL.Path = h.pathForImage(parts)
	retURL.RawQuery = vals.Encode()

	h.logger.Debug("Imagizer URL: %s", retURL.String())
	return retURL
}

func (h imagizerHandler) isPhotographerImage(id int) bool {
	pic := h.db.loadPicture(id, h.logger)
	ev := h.db.loadEvent(pic.eventID, h.logger)

	is := pic.userID == ev.ownerID
	h.logger.Debug("is photographer image? %v", is)
	return is
}

func (h imagizerHandler) getWatermarkInfo(id int, env string) watermark {
	pic := h.db.loadPicture(id, h.logger)
	pi := h.db.loadPhotographerInfo(pic.userID, h.logger)
	wm := h.db.loadWatermark(pi.id, h.logger)

	if wm.id == 0 {
		wm = watermark{
			logo:     newNullString("https://www.snapshots.com/images/icon.png"),
			disabled: false,
			alpha:    70,
			scale:    15,
			offset:   3,
			position: "bottom,right",
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

	h.logger.Debug("Watermark: %v", wm)
	return wm
}

func (h imagizerHandler) pathForImage(parts map[string]string) string {
	pictureID, err := strconv.Atoi(parts["id"])
	h.logger.HandleErr(err)

	pic := h.db.loadPicture(pictureID, h.logger)
	env, _ := parts["env"]

	var envAndUsername string

	if env == "development" {
		envAndUsername = fmt.Sprintf("%s/%s", env, parts["username"])
	} else {
		envAndUsername = env
	}

	path := fmt.Sprintf(picturePathFmt,
		BucketNames[env], envAndUsername, pictureID, pic.attachment)
	h.logger.Debug("Path for image: %s", path)
	return path
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
