package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/mantit0985/go-template/src/core/validate"
)

// Router is a wrapper around http.ServeMux that provides middleware support and route grouping.
type Router struct {
	mux                *http.ServeMux
	publicMiddlewares  []Middleware
	privateMiddlewares []Middleware
	prefix             string
	validator          validate.ValidatorAPI
}

// NewRouter creates a new Router instance and configures the global middleware chain.
func NewRouter() *Router {
	r := &Router{
		mux: http.NewServeMux(),
	}

	// Public middleware chain (bypasses Auth and Timeout)
	r.publicMiddlewares = []Middleware{
		RequestID,
		RealIP,
		Logger,
		Recoverer,
		Compress,
	}

	// Private middleware chain (added on top of public)
	r.privateMiddlewares = []Middleware{
		Auth,
		Timeout(30 * time.Second),
	}

	return r
}

// Use adds middleware to the router's private chain.
func (r *Router) Use(m ...Middleware) {
	r.privateMiddlewares = append(r.privateMiddlewares, m...)
}

// SetValidator sets the validator for the router.
func (r *Router) SetValidator(v validate.ValidatorAPI) {
	r.validator = v
}

// Handle registers a handler for the given pattern.
func (r *Router) Handle(pattern string, handler http.Handler) {
	path := strings.TrimPrefix(pattern, "/")
	fullPattern := r.prefix + path
	if !strings.HasPrefix(fullPattern, "/") {
		fullPattern = "/" + fullPattern
	}

	allMiddlewares := append([]Middleware{}, r.publicMiddlewares...)
	allMiddlewares = append(allMiddlewares, r.privateMiddlewares...)
	wrappedHandler := Chain(handler, allMiddlewares...)
	r.mux.Handle(fullPattern, wrappedHandler)
}

// PublicHandle registers a handler that bypasses secure middleware.
func (r *Router) PublicHandle(pattern string, handler http.Handler) {
	path := strings.TrimPrefix(pattern, "/")
	fullPattern := r.prefix + path
	if !strings.HasPrefix(fullPattern, "/") {
		fullPattern = "/" + fullPattern
	}

	wrappedHandler := Chain(handler, r.publicMiddlewares...)
	r.mux.Handle(fullPattern, wrappedHandler)
}

// HandleFunc registers a handler function for the given pattern.
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.Handle(pattern, http.HandlerFunc(handler))
}

// Route creates a new sub-router for a given prefix, allowing for feature isolation.
func (r *Router) Route(prefix string, fn func(sub *Router)) {
	// Ensure prefix ends with a slash for consistency
	normalizedPrefix := prefix
	if !strings.HasSuffix(normalizedPrefix, "/") {
		normalizedPrefix += "/"
	}

	sub := &Router{
		mux:                r.mux, // Use the same mux
		publicMiddlewares:  append([]Middleware{}, r.publicMiddlewares...),
		privateMiddlewares: append([]Middleware{}, r.privateMiddlewares...),
		prefix:             r.prefix + normalizedPrefix,
	}

	fn(sub)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Timeout returns a middleware that limits the request duration.
func Timeout(t time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, t, "Request Timeout")
	}
}
