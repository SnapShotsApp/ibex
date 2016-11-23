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

const re = `(?i)uploads/(?P<env>\w+)/(?P<username>\w+)?/?picture/attachment/(?P<id>\d+)/(?P<name>\w+)$`
const watermarkPathFmt = "%s/uploads/%s/%s/%d/%s"
const photographerInfoPathPart = "photographer_info/picture"
const watermarkPathPart = "watermark/logo"
const picturePathFmt = "%s/uploads/%s/picture/attachment/%d/%s"

var pathMatcher *regexp.Regexp
var imagizerHost *url.URL

type imagizerHandler struct {
	config *Config
	db     *DB
}

func init() {
	pathMatcher = regexp.MustCompile(re)
}

// Start initializes and then starts the HTTP server
func Start(c *Config) {
	db, err := NewDB(c)
	handleErr(err)

	imagizerHost, err = url.Parse(c.ImagizerHost)
	handleErr(err)

	handler := imagizerHandler{
		config: c,
		db:     db,
	}

	s := &http.Server{
		Addr:    c.BindAddr(),
		Handler: handler,
	}

	Info("Listening on %s", s.Addr)
	err = s.ListenAndServe()
	handleErr(err)
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !pathMatcher.MatchString(req.URL.Path) {
		Warn("Malformed path %s", req.URL.Path)
		http.NotFound(w, req)
		return
	}

	parts := extractPathPartsToMap(req.URL.Path)
	Debug("URL Parts: %v", parts)

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		Warn("Version not found with name %s", parts["name"])
		http.NotFound(w, req)
		return
	}
	Debug("Version found: %v", version)

	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	pic := h.db.loadPicture(pictureID)
	if pic.eventID == 0 {
		Warn("Picture not found for ID %d", pictureID)
		http.NotFound(w, req)
		return
	}

	proxy := h.imagizerURL(version, parts)
	resp, err := http.Get(proxy.String())
	defer closeQuietly(resp.Body)
	handleErr(err)

	Debug("Imagizer response: %v", resp)

	_, err = io.Copy(w, resp.Body)
	handleErr(err)
}

func (h imagizerHandler) imagizerURL(version map[string]interface{}, parts map[string]string) url.URL {
	vals := url.Values{}
	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

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
			Fatal("Unexpected type %T for %v", val, val)
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
	retURL.Scheme = imagizerHost.Scheme
	retURL.Host = imagizerHost.Host
	retURL.Path = h.pathForImage(parts)
	retURL.RawQuery = vals.Encode()

	Debug("Imagizer URL: %s", retURL.String())
	return retURL
}

func (h imagizerHandler) isPhotographerImage(id int) bool {
	pic := h.db.loadPicture(id)
	ev := h.db.loadEvent(pic.eventID)

	is := pic.userID == ev.ownerID
	Debug("is photographer image? %v", is)
	return is
}

func (h imagizerHandler) getWatermarkInfo(id int, env string) watermark {
	pic := h.db.loadPicture(id)
	pi := h.db.loadPhotographerInfo(pic.userID)
	wm := h.db.loadWatermark(pi.id)

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

	Debug("Watermark: %v", wm)
	return wm
}

func (h imagizerHandler) pathForImage(parts map[string]string) string {
	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	pic := h.db.loadPicture(pictureID)
	env, _ := parts["env"]

	var envAndUsername string

	if env == "development" {
		envAndUsername = fmt.Sprintf("%s/%s", env, parts["username"])
	} else {
		envAndUsername = env
	}

	path := fmt.Sprintf(picturePathFmt,
		BucketNames[env], envAndUsername, pictureID, pic.attachment)
	Debug("Path for image: %s", path)
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
