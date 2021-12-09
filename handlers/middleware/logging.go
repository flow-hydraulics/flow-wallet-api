package middleware

import (
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/sirupsen/logrus"
)

type responseSnooper struct {
	w      http.ResponseWriter
	status int
	size   int
	start  time.Time
}

func (l *responseSnooper) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseSnooper) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func makeSnooper(w http.ResponseWriter) (*responseSnooper, http.ResponseWriter) {
	snooper := &responseSnooper{
		w:      w,
		status: http.StatusOK, // default status
		size:   0,
		start:  time.Now(),
	}

	hooks := httpsnoop.Hooks{
		Write: func(httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return snooper.Write
		},
		WriteHeader: func(httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return snooper.WriteHeader
		},
	}

	return snooper, httpsnoop.Wrap(w, hooks)
}

func LoggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		snooper, rw := makeSnooper(rw)

		h.ServeHTTP(rw, r)

		fields := logrus.Fields{
			"method":     r.Method,
			"path":       r.RequestURI,
			"remote":     r.RemoteAddr,
			"user-agent": r.UserAgent(),
			"status":     snooper.status,
			"size":       snooper.size,
			"duration":   float64(time.Since(snooper.start).Microseconds()) / float64(1000),
		}

		logrus.WithFields(fields).Info("HTTP request")
	})
}
