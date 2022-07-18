package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/matiasinsaurralde/product-services/payment"
)

// PathType is used by parsePath and the main router to diferentiate
// all available routes
type PathType int

const (
	// PATH_ROOT state is used for the main page path:
	PATH_ROOT = iota
	// PATH_DIR state is used for paths that involve a directory:
	PATH_DIR
	// PATH_PAYMENT state is used for paths that involve a payments file:
	PATH_PAYMENT
	// PATH_ERROR state is used for all other paths that don't match the existing ones:
	PATH_ERROR
)

// Handler is the main API Handler abstraction
// Handler implements ServeHTTP in order to satisfy the http.Handler interface
type Handler struct {
	// paymentsService wraps the logic of the payments service associated with this handler
	paymentsService *payment.PaymentsService
}

// NewHandler initializes a new API handler with baseDir as the base data directory
// NewHandler returns an http.Handler and an error if the payment service initialization failed.
func NewHandler(baseDir string) (http.Handler, error) {
	paymentsService, err := payment.NewWithBaseDir(baseDir)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		paymentsService: paymentsService,
	}
	return h, nil
}

// parsePath is a helper to cleanup the URL path and extract its params
// also returns a different state for every supported route
func (h *Handler) parsePath(path string) (t PathType, params []string) {
	// Basically just get rid of all slashes and reconstruct the URL
	// with clean params:
	for _, s := range strings.Split(path, "/") {
		if s == "" {
			continue
		}
		params = append(params, s)
	}
	switch len(params) {
	case 0:
		return PATH_ROOT, nil
	case 1:
		return PATH_DIR, params
	case 2:
		return PATH_PAYMENT, params
	default:
		return PATH_ERROR, nil
	}
}

// serveNotFound is a helper that returns HTTP 404
func (h *Handler) serveNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
	w.Write([]byte("not found"))
}

// serveError is a helper that returns HTTP 500
func (h *Handler) serveError(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("server error"))
}

// ServeHTTP satisfies the http.Handler interface by implementing all the HTTP logic of the API
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathType, urlParams := h.parsePath(r.URL.Path)
	switch pathType {
	case PATH_PAYMENT:
		// Call GetPayments with all available URL params
		// In this case urlParams looks like YYYYMMDD/HHMMSS.payment
		payments, err := h.paymentsService.GetPayments(strings.Join(urlParams, "/"))
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveNotFound(w)
			return
		}
		paymentsJSON, err := json.Marshal(payments)
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveError(w)
			return
		}
		w.WriteHeader(200)
		w.Header().Add("content-type", "application/json")
		w.Write(paymentsJSON)
		return
	case PATH_ROOT:
		dirs, err := h.paymentsService.ListDirectories()
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveError(w)
			return
		}
		dirsJSON, err := json.Marshal(dirs)
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveError(w)
			return
		}
		w.WriteHeader(200)
		w.Header().Add("content-type", "application/json")
		w.Write(dirsJSON)
		return
	case PATH_DIR:
		// Call ListPayments with a single parameter, like "YYYYMMDD":
		dirs, err := h.paymentsService.ListPayments(urlParams[0])
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveNotFound(w)
			return
		}
		dirsJSON, err := json.Marshal(dirs)
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			h.serveError(w)
			return
		}
		w.WriteHeader(200)
		w.Header().Add("content-type", "application/json")
		w.Write(dirsJSON)
		return
	case PATH_ERROR:
		h.serveError(w)
		return
	}
}
