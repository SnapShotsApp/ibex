package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const re = `(?i)uploads/(?P<env>\w+)/(?P<username>\w+)?/?picture/attachment/(?P<id>\d+)/(?P<name>\w+)$`
const watermarkPathFmt = "%s/uploads/%s/photographer_info/picture/%d/%s"
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

	log.Printf("Listening on %s", s.Addr)
	err = s.ListenAndServe()
	handleErr(err)
}

func (h imagizerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !pathMatcher.MatchString(req.URL.Path) {
		log.Printf("Malformed path %s", req.URL.Path)
		http.NotFound(w, req)
		return
	}

	parts := extractPathPartsToMap(req.URL.Path)

	version, ok := h.config.versionsByName[parts["name"]]
	if !ok {
		log.Printf("Version not found with name %s", parts["name"])
		http.NotFound(w, req)
		return
	}

	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	pic := h.db.loadPicture(pictureID)
	if pic.eventID == 0 {
		log.Printf("Picture not found for ID %d", pictureID)
		http.NotFound(w, req)
		return
	}

	log.Printf("Attempting proxy for %v", version)
	proxy := h.imagizerURL(version, parts)
	resp, err := http.Get(proxy.String())
	defer closeQuietly(resp.Body)
	handleErr(err)

	_, err = io.Copy(w, resp.Body)
	handleErr(err)
}

func (h imagizerHandler) imagizerURL(version map[string]interface{}, parts map[string]string) url.URL {
	vals := url.Values{}
	pictureID, err := strconv.Atoi(parts["id"])
	handleErr(err)

	for key, val := range version {
		if key == "watermark" && val == true && h.isPhotographerImage(pictureID) {
			vals.Add("mark", h.getWatermarkURL(pictureID, parts["env"]))
			vals.Add("mark_scale", "15")
			vals.Add("mark_pos", "bottom,right")
			vals.Add("mark_offset", "3")
			vals.Add("mark_alpha", "70")
			continue
		}

		if key == "function_name" || key == "name" || key == "only_shrink_larger" {
			continue
		}

		switch val := val.(type) {
		default:
			log.Fatalf("Unexpected type %T for %v", val, val)
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

	log.Printf("Proxying to %v", retURL)

	return retURL
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

	var envAndUsername string

	if env == "development" {
		envAndUsername = fmt.Sprintf("%s/%s", env, parts["username"])
	} else {
		envAndUsername = env
	}

	return fmt.Sprintf(picturePathFmt,
		BucketNames[env], envAndUsername, pictureID, pic.attachment)
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
