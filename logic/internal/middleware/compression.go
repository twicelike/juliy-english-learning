package middleware

import (
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type brotliWriter struct {
	gin.ResponseWriter
	writer *brotli.Writer
}

func (w *brotliWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

func (w *brotliWriter) WriteString(s string) (int, error) {
	return w.writer.Write([]byte(s))
}

// Compression intelligently applies Brotli or Gzip compression based on Accept-Encoding.
func Compression() gin.HandlerFunc {
	gzipHandler := gzip.Gzip(gzip.DefaultCompression)

	return func(c *gin.Context) {
		// Do not compress if already compressed or if it's SSE/Websocket
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Next()
			return
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")

		if strings.Contains(acceptEncoding, "br") {
			c.Writer.Header().Set("Content-Encoding", "br")
			c.Writer.Header().Set("Vary", "Accept-Encoding")

			bw := brotli.NewWriterLevel(c.Writer, brotli.DefaultCompression)
			defer bw.Close()

			bwWrapper := &brotliWriter{ResponseWriter: c.Writer, writer: bw}
			c.Writer = bwWrapper
			c.Next()
			return
		}

		if strings.Contains(acceptEncoding, "gzip") {
			gzipHandler(c)
			return
		}

		c.Next()
	}
}
