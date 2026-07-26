package docsui

import (
	"embed"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"hublio/internal/platform/env"

	"github.com/gin-gonic/gin"
)

//go:embed scalar.html
var scalarFS embed.FS

const (
	defaultDocsPath = "/docs"
	specPath        = "/docs/openapi.yaml"
)

var (
	specOnce sync.Once
	specData []byte
	specErr  error
)

// Enabled reports whether interactive API docs should be mounted.
// Defaults on in development; override with ENABLE_API_DOCS=true|false.
func Enabled() bool {
	switch strings.ToLower(strings.TrimSpace(env.GetEnv("ENABLE_API_DOCS", ""))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	mode := strings.ToLower(env.GetEnv("DEVELOPMENT_MODE", "development"))
	return mode == "development" || mode == "dev" || mode == "local"
}

// Register mounts Scalar UI (Scramble-like interactive docs) and the OpenAPI YAML.
func Register(router gin.IRoutes) {
	tmpl := template.Must(template.ParseFS(scalarFS, "scalar.html"))

	router.GET(defaultDocsPath, func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(c.Writer, map[string]string{
			"Title":   "Hublio API",
			"SpecURL": specPath,
		})
	})

	router.GET(specPath, func(c *gin.Context) {
		data, err := loadOpenAPI()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "openapi spec not found; expect api/openapi/openapi.yaml at repo root",
			})
			return
		}
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
	})
}

func loadOpenAPI() ([]byte, error) {
	specOnce.Do(func() {
		for _, candidate := range openAPICandidates() {
			data, err := os.ReadFile(candidate)
			if err == nil {
				specData = data
				return
			}
			specErr = err
		}
	})
	if len(specData) > 0 {
		return specData, nil
	}
	return nil, specErr
}

func openAPICandidates() []string {
	rel := filepath.Join("api", "openapi", "openapi.yaml")

	var out []string
	seen := make(map[string]struct{})
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	// Explicit override wins (absolute path to the spec file).
	if override := strings.TrimSpace(env.GetEnv("OPENAPI_SPEC_PATH", "")); override != "" {
		add(override)
	}

	add(rel)

	// Walk up from the working directory so the spec is found whether the
	// process runs at the repo root or inside apps/api (monorepo layout).
	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, rel))
		for dir := cwd; ; {
			add(filepath.Join(dir, rel))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// Walk up from the executable directory as a fallback.
	if ex, err := os.Executable(); err == nil {
		for dir := filepath.Dir(ex); ; {
			add(filepath.Join(dir, rel))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return out
}
