package gziphandler

import (
	"compress/gzip"
	"fmt"
	"strings"
)

// Used for functional configuration.
type config struct {
	minSize      int
	level        int
	contentTypes []string
}

func (c *config) validate() error {
	if c.level != gzip.DefaultCompression && (c.level < gzip.BestSpeed || c.level > gzip.BestCompression) {
		return fmt.Errorf("invalid compression level requested: %d", c.level)
	}

	if c.minSize < 0 {
		return fmt.Errorf("minimum size must be more than zero")
	}

	return nil
}

type option func(c *config)

func MinSize(size int) option {
	return func(c *config) {
		c.minSize = size
	}
}

func CompressionLevel(level int) option {
	return func(c *config) {
		c.level = level
	}
}

func ContentTypes(types []string) option {
	return func(c *config) {
		c.contentTypes = []string{}
		for _, v := range types {
			c.contentTypes = append(c.contentTypes, strings.ToLower(v))
		}
	}
}
