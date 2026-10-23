package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yingce/backend/internal/protocol"
)

func TestAPIMartPackageInstallation(t *testing.T) {
	ids := []string{"apimart-gpt-image-2", "apimart-grok-imagine-video"}
	packages := make(map[string][]byte, len(ids))
	for _, id := range ids {
		data, err := os.ReadFile(filepath.Join("..", "..", "..", "plugin-packages", id+".yingce-plugin"))
		if err != nil {
			t.Fatal(err)
		}
		pkg, err := protocol.ParsePluginPackage(data)
		if err != nil {
			t.Fatalf("parse %s: %v", id, err)
		}
		if _, err := protocol.LoadInstalledProviders(pkg.ManifestRaw, nil); err != nil {
			t.Fatalf("load %s: %v", id, err)
		}
		packages[id] = data
	}

	t.Run("official packages reject upload replacement", func(t *testing.T) {
		center, err := newPluginRuntime(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			_, err := center.install(packages[id], id+".yingce-plugin")
			if err == nil || !strings.Contains(err.Error(), "不能通过上传覆盖") {
				t.Fatalf("official %s: expected overwrite rejection, got %v", id, err)
			}
			t.Logf("%s: %v", id, err)
		}
	})

	t.Run("external packages install and update", func(t *testing.T) {
		t.Setenv("CANVAS_OFFICIAL_PLUGIN_DIR", t.TempDir())
		center, err := newPluginRuntime(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			for range 2 {
				installed, err := center.install(packages[id], id+".yingce-plugin")
				if err != nil {
					t.Fatalf("install %s: %v", id, err)
				}
				if installed.Source != PluginOriginUploaded || installed.Status != "enabled" {
					t.Fatalf("installed %s: %#v", id, installed)
				}
			}
		}
	})
}
