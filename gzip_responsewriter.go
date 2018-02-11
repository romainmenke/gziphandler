package gziphandler

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
)

// GzipResponseWriter provides an http.ResponseWriter interface, which gzips
// bytes before writing them to the underlying response. This doesn't close the
// writers, so don't forget to do that.
// It can be configured to skip response smaller than minSize.
type GzipResponseWriter struct {
	http.ResponseWriter
	index int // Index for gzipWriterPools.
	gw    *gzip.Writer

	code int // Saves the WriteHeader value.

	minSize int    // Specifed the minimum response size to gzip. If the response length is bigger than this value, it is compressed.
	buf     []byte // Holds the first part of the write before reaching the minSize or the end of the write.

	contentTypes []string // Only compress if the response is one of these content-types. All are accepted if empty.
}

type GzipResponseWriterWithCloseNotify struct {
	*GzipResponseWriter
}

func (w GzipResponseWriterWithCloseNotify) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Write appends data to the gzip writer.
func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	// If content type is not set.
	if _, ok := w.Header()[contentType]; !ok {
		// It infer it from the uncompressed body.
		w.Header().Set(contentType, http.DetectContentType(b))
	}

	// GZIP responseWriter is initialized. Use the GZIP responseWriter.
	if w.gw != nil {
		n, err := w.gw.Write(b)
		return n, err
	}

	// Save the write into a buffer for later use in GZIP responseWriter (if content is long enough) or at close with regular responseWriter.
	// On the first write, w.buf changes from nil to a valid slice
	w.buf = append(w.buf, b...)

	// If the global writes are bigger than the minSize and we're about to write
	// a response containing a content type we want to handle, enable
	// compression.
	if len(w.buf) >= w.minSize && handleContentType(w.contentTypes, w) && w.Header().Get(contentEncoding) == "" {
		err := w.startGzip()
		if err != nil {
			return 0, err
		}
	}

	return len(b), nil
}

// startGzip initialize any GZIP specific informations.
func (w *GzipResponseWriter) startGzip() error {

	// Set the GZIP header.
	w.Header().Set(contentEncoding, "gzip")

	// if the Content-Length is already set, then calls to Write on gzip
	// will fail to set the Content-Length header since its already set
	// See: https://github.com/golang/go/issues/14975.
	w.Header().Del(contentLength)

	// Write the header to gzip response.
	if w.code != 0 {
		w.ResponseWriter.WriteHeader(w.code)
	}

	// Initialize the GZIP response.
	w.init()

	// Flush the buffer into the gzip response.
	n, err := w.gw.Write(w.buf)

	// This should never happen (per io.Writer docs), but if the write didn't
	// accept the entire buffer but returned no specific error, we have no clue
	// what's going on, so abort just to be safe.
	if err == nil && n < len(w.buf) {
		return io.ErrShortWrite
	}

	w.buf = nil
	return err
}

// WriteHeader just saves the response code until close or GZIP effective writes.
func (w *GzipResponseWriter) WriteHeader(code int) {
	if w.code == 0 {
		w.code = code
	}
}

// init graps a new gzip writer from the gzipWriterPool and writes the correct
// content encoding header.
func (w *GzipResponseWriter) init() {
	// Bytes written during ServeHTTP are redirected to this gzip writer
	// before being written to the underlying response.
	gzw := gzipWriterPools[w.index].Get().(*gzip.Writer)
	gzw.Reset(w.ResponseWriter)
	w.gw = gzw
}

// Close will close the gzip.Writer and will put it back in the gzipWriterPool.
func (w *GzipResponseWriter) Close() error {
	if w.gw == nil {
		// Gzip not trigged yet, write out regular response.
		if w.code != 0 {
			w.ResponseWriter.WriteHeader(w.code)
		}
		if w.buf != nil {
			_, writeErr := w.ResponseWriter.Write(w.buf)
			// Returns the error if any at write.
			if writeErr != nil {
				return fmt.Errorf("gziphandler: write to regular responseWriter at close gets error: %q", writeErr.Error())
			}
		}
		return nil
	}

	err := w.gw.Close()
	gzipWriterPools[w.index].Put(w.gw)
	w.gw = nil
	return err
}

// Flush flushes the underlying *gzip.Writer and then the underlying
// http.ResponseWriter if it is an http.Flusher. This makes GzipResponseWriter
// an http.Flusher.
func (w *GzipResponseWriter) Flush() {
	if w.gw == nil {
		// Only flush once startGzip has been called.
		//
		// Flush is thus a no-op until the written body
		// exceeds minSize.
		return
	}

	w.gw.Flush()

	if fw, ok := w.ResponseWriter.(http.Flusher); ok {
		fw.Flush()
	}
}

// Hijack implements http.Hijacker. If the underlying ResponseWriter is a
// Hijacker, its Hijack method is returned. Otherwise an error is returned.
func (w *GzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("http.Hijacker interface is not supported")
}

// verify Hijacker interface implementation
var _ http.Hijacker = &GzipResponseWriter{}
