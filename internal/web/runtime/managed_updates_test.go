package runtime

import (
	"context"
	"errors"
	"github.com/mhsanaei/3x-ui/v3/internal/config"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestOpsManagedUpdatesRemoteNeverContactsNode(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(http.StatusOK) }))
	defer srv.Close()
	r := NewRemote(nodeForPlainServer(t, srv, "verify", "test"), nil)
	for _, dev := range []bool{false, true} {
		if err := r.UpdatePanel(context.Background(), dev); !errors.Is(err, config.ErrBuiltinUpdatesDisabled) {
			t.Fatalf("update error: %v", err)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("update contacted a node %d times", calls.Load())
	}
}
