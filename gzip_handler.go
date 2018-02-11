package gziphandler

import (
	"compress/gzip"
	"net/http"
)

// MustNewGzipLevelHandler behaves just like NewGzipLevelHandler except that in
// an error case it panics rather than returning an error.
func MustNewGzipLevelHandler(level int) func(http.Handler) http.Handler {
	wrap, err := NewGzipLevelHandler(level)
	if err != nil {
		panic(err)
	}
	return wrap
}

// NewGzipLevelHandler returns a wrapper function (often known as middleware)
// which can be used to wrap an HTTP handler to transparently gzip the response
// body if the client supports it (via the Accept-Encoding header). Responses will
// be encoded at the given gzip compression level. An error will be returned only
// if an invalid gzip compression level is given, so if one can ensure the level
// is valid, the returned error can be safely ignored.
func NewGzipLevelHandler(level int) (func(http.Handler) http.Handler, error) {
	return NewGzipLevelAndMinSize(level, DefaultMinSize)
}

// NewGzipLevelAndMinSize behave as NewGzipLevelHandler except it let the caller
// specify the minimum size before compression.
func NewGzipLevelAndMinSize(level, minSize int) (func(http.Handler) http.Handler, error) {
	return GzipHandlerWithOpts(CompressionLevel(level), MinSize(minSize))
}

func GzipHandlerWithOpts(opts ...option) (func(http.Handler) http.Handler, error) {
	c := &config{
		level:   gzip.DefaultCompression,
		minSize: DefaultMinSize,
	}

	for _, o := range opts {
		o(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return func(h http.Handler) http.Handler {
		index := poolIndex(c.level)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(vary, acceptEncoding)
			if acceptsGzip(r) {
				gw := &GzipResponseWriter{
					ResponseWriter: w,
					index:          index,
					minSize:        c.minSize,
					contentTypes:   c.contentTypes,
				}
				defer gw.Close()

				if _, ok := w.(http.CloseNotifier); ok {
					gwcn := GzipResponseWriterWithCloseNotify{gw}
					h.ServeHTTP(gwcn, r)
				} else {
					h.ServeHTTP(gw, r)
				}

			} else {
				h.ServeHTTP(w, r)
			}
		})
	}, nil
}

// GzipHandler wraps an HTTP handler, to transparently gzip the response body if
// the client supports it (via the Accept-Encoding header). This will compress at
// the default compression level.
func GzipHandler(h http.Handler) http.Handler {
	wrapper, _ := NewGzipLevelHandler(gzip.DefaultCompression)
	return wrapper(h)
}
