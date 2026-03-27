// Package server provides the HTTP server for the Gomoku AI.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/todd/earthmover/internal/ai"
	"github.com/todd/earthmover/internal/board"
)

const maxInstances = 3

type instance struct {
	ai        *ai.AI
	sessionID string
}

// Server is the HTTP server for the Gomoku AI.
type Server struct {
	mu        sync.Mutex
	instances [maxInstances]*instance
	sessions  map[string]int // sessionID → instance index

	staticDir string
}

// New creates a new Server. staticDir is the path to the frontend files.
func New(staticDir string) *Server {
	return &Server{
		sessions:  make(map[string]int),
		staticDir: staticDir,
	}
}

// Run starts the HTTP server on the given port.
func (s *Server) Run(port int) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/start", s.handleStart)
	mux.HandleFunc("/play", s.handlePlay)
	mux.HandleFunc("/think", s.handleThink)
	mux.HandleFunc("/pass", s.handlePass)
	mux.HandleFunc("/undo", s.handleUndo)
	mux.HandleFunc("/resign", s.handleResign)
	mux.HandleFunc("/quit", s.handleQuit)
	mux.HandleFunc("/keepAlive", s.handleKeepAlive)
	mux.HandleFunc("/visualize", s.handleVisualize)
	mux.HandleFunc("/usage", s.handleUsage)

	// Static files
	mux.HandleFunc("/", s.handleStatic)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Listening on port %d (pid: %d)\n", port, os.Getpid())
	return http.ListenAndServe(addr, mux)
}

// --- Request/response helpers ---

type requestBody struct {
	SessionID string `json:"sessionId"`
	Level     int    `json:"level"`
	Rule      int    `json:"rule"`
	Row       int    `json:"row"`
	Col       int    `json:"col"`
	Times     int    `json:"times"`
}

func decodeBody(r *http.Request) (requestBody, error) {
	var body requestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, err
	}
	return body, nil
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) getInstance(sessionID string) (int, *ai.AI) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.sessions[sessionID]
	if !ok || s.instances[idx] == nil {
		return -1, nil
	}
	return idx, s.instances[idx].ai
}

// --- Handlers ---

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()

	// Check if session already has an instance
	idx, exists := s.sessions[body.SessionID]
	if !exists {
		// Find a free or expired slot
		idx = -1
		for i := 0; i < maxInstances; i++ {
			if s.instances[i] == nil {
				s.instances[i] = &instance{ai: ai.New()}
				idx = i
				break
			} else if !s.instances[i].ai.IsAlive() {
				// Reuse expired slot
				s.instances[i].ai.StopBackground()
				oldSession := s.instances[i].sessionID
				delete(s.sessions, oldSession)
				idx = i
				break
			}
		}

		if idx == -1 {
			s.mu.Unlock()
			http.Error(w, "server busy", http.StatusServiceUnavailable)
			return
		}

		s.sessions[body.SessionID] = idx
		s.instances[idx].sessionID = body.SessionID
		log.Printf("sessionID=%s instance created %d\n", body.SessionID, idx)
	}

	inst := s.instances[idx]
	s.mu.Unlock()

	inst.ai.StopBackground()
	inst.ai.Reset(body.Level, board.Rule(body.Rule))
	inst.ai.RenewLiveTime()

	w.WriteHeader(http.StatusNoContent)
	log.Printf("request path: /start => 204\n")
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.StopBackground()

	index := body.Row*board.Dimen + body.Col
	winner := em.Play(index)

	jsonResponse(w, map[string]int{"winner": winner})
	log.Printf("request path: /play => 200\n")
}

func (s *Server) handleThink(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.StopBackground()

	// Think (blocking)
	index := em.Think()

	whoWin := -1
	if index == -1 {
		em.Pass()
	} else {
		whoWin = em.Play(index)
	}

	resp := map[string]int{
		"row":    -1,
		"col":    -1,
		"winner": whoWin,
	}
	if index != -1 {
		resp["row"] = index / board.Dimen
		resp["col"] = index % board.Dimen
	}

	jsonResponse(w, resp)
	log.Printf("request path: /think => 200\n")

	// Start background thinking if game not over
	if whoWin == -1 {
		em.ThinkInBackground()
	}
}

func (s *Server) handlePass(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.StopBackground()
	em.Pass()

	w.WriteHeader(http.StatusNoContent)
	log.Printf("request path: /pass => 204\n")
}

func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.StopBackground()
	em.Undo(body.Times)

	w.WriteHeader(http.StatusNoContent)
	log.Printf("request path: /undo => 204\n")
}

func (s *Server) handleResign(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.StopBackground()

	w.WriteHeader(http.StatusNoContent)
	log.Printf("request path: /resign => 204\n")
}

func (s *Server) handleQuit(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	idx, ok := s.sessions[body.SessionID]
	if ok && s.instances[idx] != nil {
		s.instances[idx].ai.StopBackground()
		delete(s.sessions, body.SessionID)
		s.instances[idx] = nil
		log.Printf("session %s quit\n", body.SessionID)
	}
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
	log.Printf("request path: /quit => 204\n")
}

func (s *Server) handleKeepAlive(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	idx, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	em.RenewLiveTime()
	log.Printf("instance %d timestamp renewed\n", idx)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleVisualize(w http.ResponseWriter, r *http.Request) {
	body, err := decodeBody(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	_, em := s.getInstance(body.SessionID)
	if em == nil {
		http.Error(w, "invalid session", http.StatusBadRequest)
		return
	}

	treeJSON := em.GetTreeJSON()
	w.Header().Set("Content-Type", "application/json")
	if treeJSON == "" {
		w.Write([]byte("{}"))
	} else {
		w.Write([]byte(treeJSON))
	}
	log.Printf("request path: /visualize => 200\n")
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	s.mu.Lock()
	aiInfo := []map[string]any{}
	for i := 0; i < maxInstances; i++ {
		if s.instances[i] != nil {
			aiInfo = append(aiInfo, map[string]any{
				"id":        i,
				"lastAlive": s.instances[i].ai.LastAliveTime().Unix(),
			})
		}
	}
	s.mu.Unlock()

	resp := map[string]any{
		"rss":         memStats.Sys,
		"num_threads": runtime.NumGoroutine(),
		"timestamp":   time.Now().Unix(),
		"ai":          aiInfo,
	}

	jsonResponse(w, resp)
	log.Printf("request path: /usage => 200\n")
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Redirect root to index.html
	if path == "/" {
		path = "/index.html"
	}

	// Sanitize: only allow specific paths
	allowed := false
	if path == "/index.html" || path == "/dashboard.html" {
		allowed = true
	}
	if len(path) > len("/gomoku/src/") && path[:len("/gomoku/src/")] == "/gomoku/src/" {
		// Map /gomoku/src/* to static/gomoku_src/*
		if !containsDotDot(path) {
			filePath := s.staticDir + "/gomoku_src" + path[len("/gomoku/src"):]
			http.ServeFile(w, r, filePath)
			return
		}
	}

	if !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	filePath := s.staticDir + path
	http.ServeFile(w, r, filePath)
}

func containsDotDot(path string) bool {
	for i := 0; i < len(path)-1; i++ {
		if path[i] == '.' && path[i+1] == '.' {
			return true
		}
	}
	return false
}
