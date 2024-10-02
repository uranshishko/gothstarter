package handlers

import (
	"cmp"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler func(hc HandlerContext) error
type HTTPErrorHandler func(hc HandlerContext, err error)

func Make(h HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hc := NewHandlerContext(w, r)

		if err := h(hc); err != nil {
			slog.Error("HTTP handler error", "err", err, "path", r.URL.Path)
		}
	}
}

type Map map[string]interface{}

type HandlerContext interface {
	Request() *http.Request
	SetRequest(*http.Request)
	Response() http.ResponseWriter
	SetResponse(http.ResponseWriter)
	Get(string) interface{}
	Set(string, interface{})
	Delete(string)
	QueryParams() url.Values
	QueryParam(string) string
	QueryString() string
	JSON(code int, v interface{}) error
	Render(templ.Component) error
	writeContentType(string)
	setStatusCode(int)
	Redirect(code int, url string)
	Getenv(key string, defaults ...string) string
	URLParam(string) string
}

const (
	ContentTypeJSON        = "application/json"
	ContentTypeHTML        = "text/html"
	ContentTypeText        = "text/plain"
	ContentTypeXML         = "text/xml"
	ContentTypeCSS         = "text/css"
	ContentTypeJS          = "application/javascript"
	ContentTypeForm        = "application/x-www-form-urlencoded"
	ContentTypeMultipart   = "multipart/form-data"
	ContentTypeOctetStream = "application/octet-stream"
)

type handlerContext struct {
	request  *http.Request
	response http.ResponseWriter

	lock  sync.RWMutex
	store Map
}

func NewHandlerContext(w http.ResponseWriter, r *http.Request) HandlerContext {
	return &handlerContext{
		request:  r,
		response: w,

		lock:  sync.RWMutex{},
		store: make(Map),
	}
}

func (h *handlerContext) Request() *http.Request {
	return h.request
}

func (h *handlerContext) SetRequest(r *http.Request) {
	h.request = r
}

func (h *handlerContext) Response() http.ResponseWriter {
	return h.response
}

func (h *handlerContext) SetResponse(r http.ResponseWriter) {
	h.response = r
}

func (h *handlerContext) Get(key string) interface{} {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if v, ok := h.store[key]; !ok {
		return nil
	} else {
		return v
	}
}

func (h *handlerContext) Set(key string, value interface{}) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.store == nil {
		h.store = make(Map)
	}

	h.store[key] = value
}

func (h *handlerContext) Delete(key string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.store[key]; !ok {
		return
	}

	delete(h.store, key)
}

func (h *handlerContext) QueryParams() url.Values {
	return h.request.URL.Query()
}

func (h *handlerContext) QueryParam(key string) string {
	return h.QueryParams().Get(key)
}

func (h *handlerContext) QueryString() string {
	return h.request.URL.RawQuery
}

func (h *handlerContext) JSON(code int, v interface{}) error {
	h.writeContentType(ContentTypeJSON)
	h.setStatusCode(code)
	return json.NewEncoder(h.response).Encode(v)
}

func (h *handlerContext) Render(c templ.Component) error {
	h.writeContentType(ContentTypeHTML)
	return c.Render(h.request.Context(), h.response)
}

func (h *handlerContext) writeContentType(contentType string) {
	header := h.response.Header()

	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", contentType)
	}
}

func (h *handlerContext) setStatusCode(code int) {
	h.response.WriteHeader(code)
}

func (h *handlerContext) Redirect(code int, url string) {
	http.Redirect(h.response, h.request, url, code)
}

func (h *handlerContext) Getenv(key string, defaults ...string) string {
	val := os.Getenv(key)
	vals := []string{val}
	vals = append(vals, defaults...)
	return cmp.Or(vals...)
}

func (h *handlerContext) URLParam(key string) string {
	return chi.URLParam(h.Request(), key)
}
