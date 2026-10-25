package core_http_response

import "net/http"

const statusUninitialized = -1

type Writer struct {
	http.ResponseWriter
	status int
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: w, status: statusUninitialized}
}

func (w *Writer) WriteHeader(code int) {
	if w.status == statusUninitialized {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *Writer) Write(body []byte) (int, error) {
	if w.status == statusUninitialized {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func (w *Writer) Status() int {
	if w.status == statusUninitialized {
		return http.StatusOK
	}
	return w.status
}

func (w *Writer) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
