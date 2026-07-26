package docsui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOpenAPICandidatesFindsSpecFromAppsAPI ensures the spec is discoverable
// when the process runs inside apps/api (monorepo layout), not just the repo
// root. Regression test for docs 500 "openapi spec not found".
func TestOpenAPICandidatesFindsSpecFromAppsAPI(t *testing.T) {
	// Locate repo root by walking up until api/openapi/openapi.yaml exists.
	start, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	rel := filepath.Join("api", "openapi", "openapi.yaml")
	root := ""
	for dir := start; ; {
		if _, statErr := os.Stat(filepath.Join(dir, rel)); statErr == nil {
			root = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root == "" {
		t.Skip("repo openapi spec not found from test cwd")
	}

	appsAPI := filepath.Join(root, "apps", "api")
	if _, statErr := os.Stat(appsAPI); statErr != nil {
		t.Skipf("apps/api not present: %v", statErr)
	}

	t.Chdir(appsAPI)

	found := false
	for _, candidate := range openAPICandidates() {
		if _, statErr := os.Stat(candidate); statErr == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no candidate resolved to an existing spec from %s", appsAPI)
	}
}
