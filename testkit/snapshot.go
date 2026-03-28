package testkit

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func AssertMatchesGolden(t testing.TB, repoRoot, actualPath, goldenRel string) {
	t.Helper()

	goldenPath := GoldenPath(repoRoot, goldenRel)
	if UpdateGoldenEnabled() {
		if err := copyFile(actualPath, goldenPath); err != nil {
			t.Fatalf("update golden %q: %v", goldenRel, err)
		}
		return
	}
	if _, err := os.Stat(goldenPath); err != nil {
		t.Fatalf("missing golden snapshot %q; rerun with TANKBATTLE_UPDATE_GOLDEN=1 to create it", goldenPath)
	}
	match, diffImg, err := DiffPNG(goldenPath, actualPath)
	if err != nil {
		t.Fatalf("compare snapshot %q: %v", goldenRel, err)
	}
	if match {
		return
	}

	failureDir := filepath.Join(repoRoot, "testdata", "failures")
	if err := os.MkdirAll(failureDir, 0o755); err != nil {
		t.Fatalf("create failure dir: %v", err)
	}
	baseName := failureFileBase(goldenRel)
	actualOut := filepath.Join(failureDir, baseName+".actual.png")
	diffOut := filepath.Join(failureDir, baseName+".diff.png")
	if err := copyFile(actualPath, actualOut); err != nil {
		t.Fatalf("copy actual snapshot: %v", err)
	}
	if err := writePNG(diffOut, diffImg); err != nil {
		t.Fatalf("write diff snapshot: %v", err)
	}
	t.Fatalf("snapshot mismatch for %q\nactual: %s\ngolden: %s\ndiff: %s", goldenRel, actualOut, goldenPath, diffOut)
}

func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, raw, 0o644)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func failureFileBase(goldenRel string) string {
	base := filepath.ToSlash(goldenRel)
	base = filepath.Clean(base)
	base = filepath.Base(base[:len(base)-len(filepath.Ext(base))])
	dir := filepath.ToSlash(filepath.Dir(goldenRel))
	if dir == "." {
		return base
	}
	return fmt.Sprintf("%s_%s", sanitizePathFragment(dir), base)
}

func sanitizePathFragment(path string) string {
	path = filepath.ToSlash(path)
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	return replacer.Replace(path)
}
