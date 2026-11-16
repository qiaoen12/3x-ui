package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/config"
)

func TestOpsManagedUpdatesRejectBeforeSideEffects(t *testing.T) {
	setupSettingTestDB(t)
	dir := t.TempDir()
	t.Setenv("XUI_BIN_FOLDER", dir)
	asset := filepath.Join(dir, "geoip.dat")
	if err := os.WriteFile(asset, []byte("existing rules"), 0600); err != nil {
		t.Fatal(err)
	}
	var downloads atomic.Int32
	original := http.DefaultTransport
	transport := original.(*http.Transport).Clone()
	transport.DialContext = func(context.Context, string, string) (net.Conn, error) {
		downloads.Add(1)
		return nil, errors.New("unexpected download")
	}
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = original; transport.CloseIdleConnections() })
	var restarted bool
	restartStub(t, &restarted)
	s := &ServerService{}
	for _, call := range []func() error{
		func() error { return s.UpdateXray("v26.10.10") },
		func() error { return s.UpdateXray("invalid-version") },
		func() error { return s.UpdateGeofile("") },
		func() error { return s.UpdateGeofile("geoip.dat") },
		func() error { return s.UpdateGeofile("../../etc/passwd") },
		func() error { _, err := (&NodeService{}).UpdatePanels([]int{1, 2}, false); return err },
		func() error { _, err := (&NodeService{}).UpdatePanels([]int{1}, true); return err },
		func() error { return (&SettingService{}).SetDevChannelEnable(true) },
	} {
		if err := call(); !errors.Is(err, config.ErrBuiltinUpdatesDisabled) {
			t.Fatalf("error = %v", err)
		}
	}
	if downloads.Load() != 0 || restarted {
		t.Fatalf("downloads=%d restart=%v", downloads.Load(), restarted)
	}
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 {
		t.Fatalf("asset directory changed: %v %v", files, err)
	}
	got, err := os.ReadFile(asset)
	if err != nil || string(got) != "existing rules" {
		t.Fatalf("asset changed: %q %v", got, err)
	}
	enabled, err := (&SettingService{}).GetDevChannelEnable()
	if err != nil || enabled {
		t.Fatalf("channel changed: %v %v", enabled, err)
	}
}

func TestOpsManagedUpdatesRejectNativeGeodataWithoutSaving(t *testing.T) {
	setupSettingTestDB(t)
	s := &XraySettingService{}
	before, err := s.GetXrayConfigTemplate()
	if err != nil {
		t.Fatal(err)
	}
	for _, template := range []string{
		`{"geodata":{"assets":[{"url":"https://example.invalid/geoip.dat","file":"geoip.dat"}]}}`,
		`{"xraySetting":{"geodata":{"assets":[{"url":"https://example.invalid/custom.dat","file":"custom.dat"}]}}}`,
	} {
		if err := s.SaveXraySetting(template); !errors.Is(err, config.ErrBuiltinUpdatesDisabled) {
			t.Fatalf("save error = %v", err)
		}
	}
	after, err := s.GetXrayConfigTemplate()
	if err != nil || before != after {
		t.Fatal("rejected request changed the stored template")
	}
	if err := s.SaveXraySetting(`{"geodata":{"assets":[]},"inbounds":[],"outbounds":[]}`); err != nil {
		t.Fatalf("ordinary config blocked: %v", err)
	}
}

func TestOpsManagedUpdatesOmitRestoredGeodataAtRuntime(t *testing.T) {
	setupSettingTestDB(t)
	template := `{"geodata":{"assets":[{"url":"https://example.invalid/geoip.dat","file":"geoip.dat"}]},"log":{"loglevel":"warning"},"inbounds":[],"outbounds":[]}`
	s := &SettingService{}
	if err := s.saveSetting("xrayTemplateConfig", template); err != nil {
		t.Fatal(err)
	}
	cfg, err := (&XrayService{}).GetXrayConfig()
	if err != nil {
		t.Fatalf("runtime config remains usable: %v", err)
	}
	if len(cfg.Geodata) != 0 {
		t.Fatalf("native downloader remains active: %s", cfg.Geodata)
	}
	stored, err := s.GetXrayConfigTemplate()
	if err != nil || stored != template {
		t.Fatal("runtime generation changed the stored template")
	}
}
