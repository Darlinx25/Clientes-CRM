package version

import (
	"os"
	"strings"
	"testing"
)

func TestFallbackMatchesVersionFile(t *testing.T) {
	raw, err := os.ReadFile("../../VERSION")
	if err != nil {
		t.Fatalf("read repository VERSION: %v", err)
	}
	want := strings.TrimSpace(string(raw))
	if Version != want {
		t.Fatalf("version.Version is %q but VERSION is %q; keep the fallback in sync", Version, want)
	}
}
