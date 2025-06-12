package http

import (
	"encoding/json"
	"fmt"
	"github.com/ljhe/scream/3rd/logrus"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var prefix = "/api"
var Router = map[string]http.HandlerFunc{
	"/": WithParams(http.MethodGet, helloHandler),
}

type RequestParams struct {
	QueryParams  map[string][]string
	PostParams   map[string][]string
	HeaderParams map[string][]string
	JSONBody     map[string]interface{}
}

func ExtractParams(r *http.Request) *RequestParams {
	params := &RequestParams{
		QueryParams:  make(map[string][]string),
		PostParams:   make(map[string][]string),
		HeaderParams: r.Header,
		JSONBody:     make(map[string]interface{}),
	}

	// 通用参数解析（GET 或 application/x-www-form-urlencoded）
	r.ParseForm()
	params.QueryParams = r.Form

	// 对于表单 POST 单独提取 PostForm
	if r.Method == http.MethodPost {
		params.PostParams = r.PostForm
	}

	// 对于 JSON body
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			json.Unmarshal(body, &params.JSONBody)
		}
		// r.Body 会被消耗，若后续仍需使用，则需保存或复用
		r.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	return params
}

type HandlerWithParams func(http.ResponseWriter, *http.Request, *RequestParams)

func WithParams(method string, handler HandlerWithParams) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if method != r.Method {
			http.Error(w, fmt.Sprintf("Method not allowed: expected %s", method), http.StatusMethodNotAllowed)
			return
		}
		params := ExtractParams(r)
		//printParams(w, r, params)
		handler(w, r, params)
		logrus.Log(def.LogsSystem).Infof("[%s] %s %s %s -> %d - %s", start.Format("15:04:05"), r.Method, r.URL.String(),
			r.Header.Get("Content-Type"), http.StatusOK, time.Since(start))
	}
}

func printParams(w http.ResponseWriter, r *http.Request, params *RequestParams) {
	log.Println("--- Header Parameters ---")
	for key, values := range params.HeaderParams {
		for _, value := range values {
			log.Printf("%s: %s\n", key, value)
		}
	}

	log.Println("--- Query / Form Parameters ---")
	for key, values := range params.QueryParams {
		for _, value := range values {
			log.Printf("%s = %s\n", key, value)
		}
	}

	log.Println("--- POST Form Parameters ---")
	for key, values := range params.PostParams {
		for _, value := range values {
			log.Printf("%s = %s\n", key, value)
		}
	}

	log.Println("--- JSON Body Parameters ---")
	for key, value := range params.JSONBody {
		log.Printf("%s = %v\n", key, value)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request, p *RequestParams) {
	fmt.Fprintf(w, "Hello World")
}
