package router

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"net/http"
	"strings"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

type etagResponseWriter struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (w *etagResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *etagResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

// ETagMiddleware 针对 GET/HEAD 请求根据内容自动生成 ETag 并支持 304 Not Modified 协商缓存
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer func() {
			// 若 buffer 未异常膨胀（超过 1MB），放回复用池；否则让 GC 回收避免常驻过大内存
			if buf.Cap() <= 1024*1024 {
				bufferPool.Put(buf)
			}
		}()

		rec := &etagResponseWriter{
			ResponseWriter: w,
			buf:            buf,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		if rec.status != http.StatusOK || rec.buf.Len() == 0 {
			if rec.status != 0 {
				w.WriteHeader(rec.status)
			}
			_, _ = w.Write(rec.buf.Bytes())
			return
		}

		bodyBytes := rec.buf.Bytes()
		crc := crc32.ChecksumIEEE(bodyBytes)
		etag := fmt.Sprintf("\"%08x-%d\"", crc, len(bodyBytes))

		w.Header().Set("ETag", etag)

		if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
			if strings.Contains(ifNoneMatch, etag) || ifNoneMatch == "*" {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		w.WriteHeader(rec.status)
		_, _ = w.Write(bodyBytes)
	})
}
