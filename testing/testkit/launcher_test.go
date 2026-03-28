package testkit

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestConfiguredE2EBinaryResolvesRelativePath(t *testing.T) {
	rootDir := t.TempDir()
	expected := filepath.Join(rootDir, "bin", "tankbattle.exe")
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(expected, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv("TANKBATTLE_E2E_BINARY", filepath.Join("bin", "tankbattle.exe"))

	got, ok, err := configuredE2EBinary(rootDir)
	if err != nil {
		t.Fatalf("configuredE2EBinary returned error: %v", err)
	}
	if !ok {
		t.Fatal("configuredE2EBinary should report configured binary")
	}
	if got != expected {
		t.Fatalf("configuredE2EBinary = %q, want %q", got, expected)
	}
}

func TestConfiguredE2EBinaryDoesNotReuseRepoArtifactsByDefault(t *testing.T) {
	rootDir := t.TempDir()
	for _, name := range []string{"tankbattle_gui.exe", "tankbattle.exe"} {
		path := filepath.Join(rootDir, name)
		if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Setenv("TANKBATTLE_E2E_BINARY", "")

	got, ok, err := configuredE2EBinary(rootDir)
	if err != nil {
		t.Fatalf("configuredE2EBinary returned error: %v", err)
	}
	if ok {
		t.Fatalf("configuredE2EBinary should not reuse default repo binaries, got %q", got)
	}
}

func TestConfiguredE2EBinaryRejectsMissingPath(t *testing.T) {
	rootDir := t.TempDir()
	t.Setenv("TANKBATTLE_E2E_BINARY", "missing.exe")

	_, ok, err := configuredE2EBinary(rootDir)
	if err == nil {
		t.Fatal("configuredE2EBinary should fail for a missing configured binary")
	}
	if ok {
		t.Fatal("configuredE2EBinary should not report a usable binary for missing paths")
	}
}

func TestSessionCloseReturnsExitedProcessError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("process exit behavior is asserted on Windows")
	}

	session := startTestSession(t, "exit 7")
	waitForExitEvent(t, session)

	err := session.Close()
	if err == nil {
		t.Fatal("Close should report unexpected child process exits")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Close error = %T, want *exec.ExitError", err)
	}
}

func TestSessionCloseSuppressesIntentionalKillError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("process exit behavior is asserted on Windows")
	}

	session := startTestSession(t, "Start-Sleep -Seconds 30")

	if err := session.Close(); err != nil {
		t.Fatalf("Close returned error for intentional kill: %v", err)
	}
}

func startTestSession(t *testing.T, script string) *Session {
	t.Helper()

	cmd := exec.Command("powershell", "-Command", script)
	session := &Session{
		cmd:    cmd,
		exitCh: make(chan error, 1),
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start process: %v", err)
	}
	go func() {
		session.exitCh <- cmd.Wait()
	}()
	return session
}

func waitForExitEvent(t *testing.T, session *Session) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-session.exitCh:
			session.exitCh <- err
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	t.Fatal("timed out waiting for child process to exit")
}
