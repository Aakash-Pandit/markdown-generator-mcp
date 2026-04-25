package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/parser"
)

func Write(content, title, outputDir, filename string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", outputDir, err)
	}
	name := resolveFilename(title, filename)
	path := filepath.Join(outputDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file %s: %w", path, err)
	}
	return path, nil
}

func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".md") || strings.HasSuffix(e.Name(), ".pdf")) {
			info, _ := e.Info()
			results = append(results, fmt.Sprintf("%s (%d bytes)", e.Name(), info.Size()))
		}
	}
	return results, nil
}

func resolveFilename(title, override string) string {
	if override != "" {
		return sanitize(override) + ".md"
	}
	if title != "" {
		slug := parser.Slugify(title)
		if slug != "" {
			return slug + ".md"
		}
	}
	return "doc-" + time.Now().Format("20060102-150405") + ".md"
}

func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
