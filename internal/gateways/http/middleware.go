package http

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type bufferedResponseWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	w.buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

func ContentLengthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		brw := &bufferedResponseWriter{
			ResponseWriter: c.Writer,
			buffer:         bytes.NewBuffer(nil),
		}
		c.Writer = brw
		c.Next()
		if brw.Header().Get("Content-Length") != "" || brw.buffer.Len() == 0 {
			return
		}
		if c.Request.Method == http.MethodHead {
			contentLength := strconv.Itoa(brw.buffer.Len())
			c.Header("Content-Length", contentLength)
		}
	}
}
