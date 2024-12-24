package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/bridge-swap/internal/service"
)

type Server struct {
	bridge *service.BridgeService
	router *mux.Router
}

func NewServer(bridge *service.BridgeService) *Server {
	s := &Server{
		bridge: bridge,
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/swap", s.handleInitiateSwap).Methods("POST")
	s.router.HandleFunc("/api/swap/{requestId}", s.handleGetSwapStatus).Methods("GET")
	s.router.HandleFunc("/api/queue/status", s.handleGetQueueStatus).Methods("GET")
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleInitiateSwap(w http.ResponseWriter, r *http.Request) {
	var req service.SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.bridge.InitiateSwap(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetSwapStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestId := vars["requestId"]

	status, err := s.bridge.GetSwapStatus(r.Context(), requestId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleGetQueueStatus(w http.ResponseWriter, r *http.Request) {
	status := s.bridge.GetQueueStatus(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
