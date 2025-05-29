package http

import (
	"github.com/ljhe/scream/core/iface"
	"net/http"
)

var Server *httpAcceptor

type httpAcceptor struct{}

func (h *httpAcceptor) Stop() {
	//TODO implement me
	panic("implement me")
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

	handler := withCORS(mux)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
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
