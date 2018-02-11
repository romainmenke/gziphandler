package gziphandler

import (
	"net/http"
	"testing"

	"github.com/fd/httpmiddlewarevet"
)

func Test(t *testing.T) {
	httpmiddlewarevet.Vet(t, func(h http.Handler) http.Handler {
		w, err := GzipHandlerWithOpts()
		if err != nil {
			t.Fatal(err)
		}
		return w(h)
	})
}
