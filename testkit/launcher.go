package testkit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type Session struct {
	RootDir     string
	RunDir      string
	SnapshotDir string
	BaseURL     string
	Client      *Client

	cmd    *exec.Cmd
	exitCh chan error
	stdout bytes.Buffer
	stderr bytes.Buffer
}

type LaunchOptions struct {
	StartupTimeout time.Duration
}

var (
	builtBinaryOnce sync.Once
	builtBinaryPath string
	builtBinaryErr  error
)

func StartSession(t testing.TB, opts LaunchOptions) *Session {
	t.Helper()

	requireE2E(t)

	rootDir, err := findRepoRoot()
	if err != nil {
		t.Fatalf("find repo root: %v", err)
	}
	artifactsDir := filepath.Join(rootDir, ".tmp_test_artifacts")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("create artifacts dir: %v", err)
	}
	runDir, err := os.MkdirTemp(artifactsDir, "e2e-")
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}
	snapshotDir := filepath.Join(runDir, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("create snapshot dir: %v", err)
	}
	binaryPath, err := ensureBuiltBinary(rootDir)
	if err != nil {
		t.Fatalf("build e2e binary: %v", err)
	}
	addr, err := allocateLocalAddr()
	if err != nil {
		t.Fatalf("allocate debug api addr: %v", err)
	}

	cmd := exec.Command(binaryPath)
	cmd.Dir = rootDir
	cmd.Env = append(os.Environ(), "TANKBATTLE_DEBUG_API_ADDR="+addr)

	session := &Session{
		RootDir:     rootDir,
		RunDir:      runDir,
		SnapshotDir: snapshotDir,
		BaseURL:     "http://" + addr,
		Client:      NewClient("http://" + addr),
		cmd:         cmd,
		exitCh:      make(chan error, 1),
	}
	cmd.Stdout = &session.stdout
	cmd.Stderr = &session.stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start debug game: %v", err)
	}
	go func() {
		session.exitCh <- cmd.Wait()
	}()

	t.Cleanup(func() {
		if err := session.Close(); err != nil && !t.Failed() {
			t.Fatalf("close debug game: %v", err)
		}
		if t.Failed() {
			t.Logf("debug game stdout:\n%s", strings.TrimSpace(session.stdout.String()))
			t.Logf("debug game stderr:\n%s", strings.TrimSpace(session.stderr.String()))
		}
	})

	timeout := opts.StartupTimeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := session.waitForReady(ctx); err != nil {
		t.Fatalf("wait for debug api: %v", err)
	}

	return session
}

func (s *Session) Close() error {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	select {
	case err := <-s.exitCh:
		return normalizeExitErr(err)
	default:
	}
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	select {
	case err := <-s.exitCh:
		return normalizeExitErr(err)
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timed out waiting for debug game to exit")
	}
}

func (s *Session) waitForReady(ctx context.Context) error {
	for {
		if _, err := s.Client.State(); err == nil {
			return nil
		}
		select {
		case err := <-s.exitCh:
			return fmt.Errorf("debug game exited before api became ready: %w; stderr=%q", err, strings.TrimSpace(s.stderr.String()))
		case <-ctx.Done():
			return fmt.Errorf("debug api did not become ready before deadline: %w; stderr=%q", ctx.Err(), strings.TrimSpace(s.stderr.String()))
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func requireE2E(t testing.TB) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("debug api e2e suite only runs on Windows")
	}
	if os.Getenv("TANKBATTLE_E2E") != "1" {
		t.Skip("set TANKBATTLE_E2E=1 to run debug api e2e suites")
	}
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found from %q", wd)
		}
		wd = parent
	}
}

func ensureBuiltBinary(rootDir string) (string, error) {
	if existing, ok := existingE2EBinary(rootDir); ok {
		return existing, nil
	}
	builtBinaryOnce.Do(func() {
		binDir := filepath.Join(rootDir, ".tmp_test_artifacts", "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			builtBinaryErr = err
			return
		}
		builtBinaryPath = filepath.Join(binDir, "tankbattle-e2e.exe")
		cmd := exec.Command("go", "build", "-mod=readonly", "-o", builtBinaryPath, "./cmd/tankbattle")
		cmd.Dir = rootDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			builtBinaryErr = fmt.Errorf("%w\n%s", err, strings.TrimSpace(string(output)))
			return
		}
	})
	return builtBinaryPath, builtBinaryErr
}

func existingE2EBinary(rootDir string) (string, bool) {
	if candidate := strings.TrimSpace(os.Getenv("TANKBATTLE_E2E_BINARY")); candidate != "" {
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(rootDir, candidate)
		}
		if fileExists(candidate) {
			return candidate, true
		}
	}
	for _, candidate := range []string{
		filepath.Join(rootDir, "tankbattle_gui.exe"),
		filepath.Join(rootDir, "tankbattle.exe"),
	} {
		if fileExists(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func allocateLocalAddr() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer ln.Close()
	return ln.Addr().String(), nil
}

func normalizeExitErr(err error) error {
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if strings.Contains(err.Error(), "killed") || strings.Contains(err.Error(), "terminated") {
		return nil
	}
	if errors.As(err, &exitErr) && exitErr != nil {
		return nil
	}
	return err
}
