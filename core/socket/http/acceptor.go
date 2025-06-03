package http

import (
	"errors"
	"github.com/ljhe/scream/core/iface"
	"log"
	"net/http"
)

type httpAcceptor struct {
	server *http.Server
}

func NewHttpServer() *httpAcceptor {
	return &httpAcceptor{}
}

func (h *httpAcceptor) Stop() {
	err := h.server.Close()
	if err != nil {
		panic(err)
	}
	log.Println("http acceptor stopped success.")
}

func (h *httpAcceptor) GetTyp() string {
	//TODO implement me
	panic("implement me")
}

func (h *httpAcceptor) Start() iface.INetNode {
	mux := http.NewServeMux()
	for k, v := range Router {
		mux.HandleFunc(prefix+k, v)
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: withCORS(mux),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	h.server = server
	return h
}

// CORS 中间件处理函数
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
