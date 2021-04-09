package web

import (
	"net/http"
	"testing"
)

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello server"))
}

func TestWebService(t *testing.T) {
	srv := NewService()
	srv.Init()
	srv.Handle("/", &handler{})
	err := srv.Run()
	if err != nil {
		t.Fatal(err)
	}
}
