package bytego

import (
	"io"
	"log"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	Size() int
	WriteString(string) (int, error)
	Committed() bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
	}
}

type responseWriter struct {
	http.ResponseWriter
	size      int
	status    int
	committed bool
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.writeStatusrCheck()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) WriteHeader(code int) {
	if w.committed {
		log.Println("writer already commited!")
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(w.status)
	w.committed = true
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.writeStatusrCheck()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Committed() bool {
	return w.committed
}

func (w *responseWriter) writeStatusrCheck() {
	if !w.committed {
		if w.status == 0 {
			w.status = http.StatusOK
		}
		w.WriteHeader(w.status)
	}
}
