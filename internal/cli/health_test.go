package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthResponses(t *testing.T) {
	for _, code := range []int{200, 503} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(code) }))
		err := healthAt(context.Background(), srv.URL)
		srv.Close()
		if (err == nil) != (code == 200) {
			t.Fatalf("%d: %v", code, err)
		}
	}
	if healthAt(context.Background(), "://") == nil {
		t.Fatal("invalid endpoint")
	}
}
