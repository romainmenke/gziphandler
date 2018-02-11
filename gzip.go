package gziphandler

import (
	"net/http"
	"strings"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	contentType     = "Content-Type"
	contentLength   = "Content-Length"
)

const (
	// DefaultQValue is the default qvalue to assign to an encoding if no explicit qvalue is set.
	// This is actually kind of ambiguous in RFC 2616, so hopefully it's correct.
	// The examples seem to indicate that it is.
	DefaultQValue = 1.0

	// DefaultMinSize defines the minimum size to reach to enable compression.
	// It's 512 bytes.
	DefaultMinSize = 512
)

// acceptsGzip returns true if the given HTTP request indicates that it will
// accept a gzipped response.
func acceptsGzip(r *http.Request) bool {
	acceptedEncodings, _ := parseEncodings(r.Header.Get(acceptEncoding))
	return acceptedEncodings["gzip"] > 0.0
}

// returns true if we've been configured to compress the specific content type.
func handleContentType(contentTypes []string, w http.ResponseWriter) bool {
	// If contentTypes is empty we handle all content types.
	if len(contentTypes) == 0 {
		return true
	}

	ct := strings.ToLower(w.Header().Get(contentType))
	for _, c := range contentTypes {
		if c == ct {
			return true
		}
	}

	return false
}
