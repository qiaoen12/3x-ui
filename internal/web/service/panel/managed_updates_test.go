package panel

import (
	"errors"
	"github.com/mhsanaei/3x-ui/v3/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestOpsManagedUpdatesPanelNeverLaunchesUpdater(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dir)
	t.Setenv("XUI_MAIN_FOLDER", dir)
	t.Setenv("XUI_SERVICE", dir)
	t.Setenv("TMPDIR", dir)
	before, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := &PanelService{}
	for _, call := range []func() (int64, error){s.StartUpdate, func() (int64, error) { return s.StartUpdateChannel(false) }, func() (int64, error) { return s.StartUpdateChannel(true) }} {
		id, err := call()
		if id != 0 || !errors.Is(err, config.ErrBuiltinUpdatesDisabled) {
			t.Fatalf("run=%d error=%v", id, err)
		}
	}
	after, err := os.ReadDir(dir)
	if err != nil || len(before) != len(after) {
		t.Fatalf("updater wrote files under %s", filepath.Clean(dir))
	}
	updateMu.Lock()
	defer updateMu.Unlock()
	if updateRunning {
		t.Fatal("disabled update acquired a process slot")
	}
}
