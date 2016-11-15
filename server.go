package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
)

const re = `(?i)uploads/(?P<env>\w+)/picture/attachment/(?P<id>\d+)/(?P<name>\w+)$`
const watermarkPathFmt = "%s/uploads/%s/photographer_info/picture/%d/%s"
const picturePathFmt = "%s/uploads/%s/picture/attachment/%d/%s"

var pathMatcher *regexp.Regexp

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

	handler := imagizerHandler{
		config: c,
		db:     db,
	}

	s := &http.Server{
		Addr:    c.BindAddr(),
		Handler: handler,
	}

	log.Printf("Listening on %s", s.Addr)
	err = s.ListenAndServe()
	handleErr(err)
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !pathMatcher.MatchString(req.URL.Path) {
		http.NotFound(w, req)
		return
	}

	parts := extractPathPartsToMap(req.URL.Path)

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		http.NotFound(w, req)
		return
	}

	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	pic := h.db.loadPicture(pictureID)
	if pic.eventID == 0 {
		http.NotFound(w, req)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: h.proxyToImagizer(version, parts),
	}

	proxy.ServeHTTP(w, req)
}

func (h imagizerHandler) proxyToImagizer(version map[string]interface{}, parts map[string]string) func(*http.Request) {
	vals := url.Values{}
	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	for key, val := range version {
		if key == "watermark" && h.isPhotographerImage(pictureID) {
			vals.Add("mark", h.getWatermarkURL(pictureID, parts["env"]))
			vals.Add("mark_scale", "75")
			vals.Add("mark_pos", "bottom,right")
			vals.Add("mark_offset", "3")
			vals.Add("mark_alpha", "70")
			continue
		}

		vals.Add(key, val.(string))
	}

	return func(req *http.Request) {
		imagizerHost, err := url.Parse(h.config.ImagizerHost)
		handleErr(err)

		req.URL.Scheme = imagizerHost.Scheme
		req.URL.Host = imagizerHost.Host
		req.URL.Path = h.pathForImage(parts)
		req.URL.RawQuery = vals.Encode()

		log.Printf("Proxying to %s", req.URL.String())
	}
}

func (h imagizerHandler) isPhotographerImage(id int) bool {
	pic := h.db.loadPicture(id)
	ev := h.db.loadEvent(pic.eventID)

	return pic.userID == ev.ownerID
}

func (h imagizerHandler) getWatermarkURL(id int, env string) string {
	pic := h.db.loadPicture(id)
	pi := h.db.loadPhotographerInfo(pic.userID)

	if pi.picture.Valid {
		return fmt.Sprintf(watermarkPathFmt, h.config.CDNHost, env, id, pi.picture)
	}

	return "https://snapshots.com/images/icon.png"
}

func (h imagizerHandler) pathForImage(parts map[string]string) string {
	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	pic := h.db.loadPicture(pictureID)
	env, _ := parts["env"]

	return fmt.Sprintf(picturePathFmt,
		BucketNames[env], env, pictureID, pic.attachment)
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
