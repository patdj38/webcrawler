package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"log"
	"net"
	"net/http"
	"runtime/debug"

	// import profiling package
	_ "net/http/pprof"
)

type Error struct {
	// ID of the error
	ID string `json:"id"`
	// HTTP status code
	Status int `json:"status"`
	// HTTP status code message
	Title string `json:"title"`
	// Detailled error message
	Detail string `json:"detail"`
}

type Errors struct {
	Errors []*Error `json:"errors"`
}

type router struct {
	*httprouter.Router
}

var notacceptableError = &Error{ID: "not_acceptable", Status: 406, Title: "Not Acceptable", Detail: "Accept header must be set correctly"}

// ContextKey Key type for context
type ContextKey string

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
	r.PUT(path, wrapHandler(handler))
}

func (r *router) Patch(path string, handler http.Handler) {
	r.PATCH(path, wrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
	r.DELETE(path, wrapHandler(handler))
}

func (r *router) Head(path string, handler http.Handler) {
	r.HEAD(path, wrapHandler(handler))
}

type contextKeyT string

var contextParamsKey = contextKeyT("params")

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, contextParamsKey, ps)))
	}
}

func newRouter() *router {
	return &router{httprouter.New()}
}

// A Server is an HTTP server that runs the FastML Engine-API REST API
type Server struct {
	router   *router
	listener net.Listener
}

// NewServer create a Server to serve the REST API
func NewServer(shutdownCh chan struct{}) (*Server, error) {
	ip := net.ParseIP("0.0.0.0")
	addr := &net.TCPAddr{IP: ip, Port: 8900}
	listener, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to bind on %s", addr)
	}

	httpServer := &Server{
		router:   newRouter(),
		listener: listener,
	}

	err = httpServer.registerHandlers()
	if err != nil {
		return nil, err
	}
	log.Printf("Starting HTTPServer on address %s", listener.Addr())

	go http.Serve(httpServer.listener, httpServer.router)

	return httpServer, nil
}

func acceptHandler(cType string) func(http.Handler) http.Handler {
	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Accept") != cType {
				writeError(w, r, notacceptableError)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}

	return m
}
func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Errorf("panic: %s", err)
				fmt.Errorf("The corresponding stack is\n%s", debug.Stack())
				rerr := &Error{ID: "internal_server_error", Status: 500, Title: "Internal Server Error", Detail: fmt.Sprint(err)}
				writeError(w, r, rerr)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) registerHandlers() error {

	var commonHandlers alice.Chain
	commonHandlers = alice.New(recoverHandler)

	//webcrawler
	// s.router.Post("/webcrawler", commonHandlers.ThenFunc(s.createExperimentHandler))
	s.router.Get("/webcrawler", commonHandlers.Append(acceptHandler("application/json")).ThenFunc(s.webCrawler))
	return nil
}

func writeError(w http.ResponseWriter, r *http.Request, err *Error) {
	fmt.Errorf(err.Detail)
	w.WriteHeader(err.Status)
	encodeJSONResponse(w, r, Errors{Errors: []*Error{err}})
}

func encodeJSONResponse(w http.ResponseWriter, r *http.Request, resp interface{}) {

	jEnc := json.NewEncoder(w)
	if _, ok := r.URL.Query()["pretty"]; ok {
		jEnc.SetIndent("", "  ")
	}
	w.Header().Set("Content-Type", "application/json")
	jEnc.Encode(resp)
}
