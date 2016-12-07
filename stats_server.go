package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type configHandler struct {
	config *Config
	logger ILogger
}

func (h configHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.logger.CloseQuietly(req.Body)
	h.logger.Debug("Request for /config")

	body, err := json.Marshal(h.config)

	if err != nil {
		http.Error(w, fmt.Sprintf("ERROR: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(body[:]))
}

type statType int

const (
	// StatBadRequest is a const for the BadRequest stat
	StatBadRequest statType = iota
	// StatTimeout is a const for the BadRequest stat
	StatTimeout
	// StatServedPicture is a const for the BadRequest stat
	StatServedPicture
)

type stat struct {
	T       statType
	Payload string
}

// Stats serves as a receptical of server statistics
type Stats struct {
	Started        time.Time         `json:"started"`
	BadRequests    uint64            `json:"bad_requests"`
	Timeouts       uint64            `json:"timeouts"`
	TotalServed    uint64            `json:"total_served"`
	TotalByVersion map[string]uint64 `json:"total_by_version"`
	statsChan      chan *stat
	logger         ILogger
}

// NewStats instantiates and returns a new stats handler
func NewStats(logger ILogger) *Stats {
	s := Stats{
		BadRequests: 0,
		Timeouts:    0,
		TotalServed: 0,
		logger:      logger,
	}
	s.TotalByVersion = make(map[string]uint64)
	s.statsChan = make(chan *stat, 10)

	return &s
}

// Listen runs the loop to handle stats collection
func (s *Stats) Listen() {
	for {
		st := <-s.statsChan
		s.logger.Debug("Incoming stat: %v", st)

		switch st.T {
		case StatBadRequest:
			s.BadRequests++
		case StatTimeout:
			s.Timeouts++
		case StatServedPicture:
			s.TotalServed++
			s.TotalByVersion[st.Payload]++
		default:
			s.logger.Warn("Unknown stat: %v", st)
		}
	}
}

func (s *Stats) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.logger.CloseQuietly(req.Body)
	s.logger.Debug("Request for /stats")

	body, err := json.Marshal(s)

	if err != nil {
		http.Error(w, fmt.Sprintf("ERROR: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(body[:]))
}

// Start starts the stats server on the specified port and starts listening for stats
func (s *Stats) Start(config *Config) {
	for _, name := range config.VersionNames() {
		s.TotalByVersion[name] = 0
	}
	s.Started = time.Now()
	go s.Listen()

	mux := http.NewServeMux()
	mux.Handle("/config", configHandler{config, s.logger})
	mux.Handle("/stats", s)

	server := &http.Server{
		Addr:         config.StatsServer.BindAddr(),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	s.logger.Info("Stats server listening on %s", server.Addr)
	err := server.ListenAndServe()
	s.logger.HandleErr(err)
}
