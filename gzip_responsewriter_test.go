package gziphandler

import (
	"net/http/httptest"
)

type responseRecorder struct {
	*httptest.ResponseRecorder
}

func (w *responseRecorder) CloseNotify() <-chan bool {
	return make(chan bool, 1000)
}

func newRecorder() *responseRecorder {
	return &responseRecorder{httptest.NewRecorder()}
}
