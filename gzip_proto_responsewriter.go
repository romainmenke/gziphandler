package gziphandler

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type h1 struct{ *GzipResponseWriter }

func (w *h1) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *h1) ReadFrom(reader io.Reader) (int64, error) {
	return io.Copy(w.GzipResponseWriter, reader)
}

type h2 struct{ *GzipResponseWriter }
