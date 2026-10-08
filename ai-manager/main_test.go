package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchFileResumesAndValidatesGGUF(t *testing.T) {
	content := append([]byte("GGUF"), []byte("tiny model fixture")...)
	sum := sha256.Sum256(content)
	var rangeSeen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeSeen = r.Header.Get("Range")
		if rangeSeen != "bytes=4-" {
			t.Errorf("Range = %q", rangeSeen)
		}
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(content[4:])
	}))
	defer server.Close()

	root := t.TempDir()
	spec := modelSpec{ID: "test", Files: []fileSpec{{Name: "test.gguf", URL: server.URL, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(content))}}}
	if err := os.MkdirAll(filepath.Join(root, spec.ID), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, spec.ID, "test.gguf.partial"), content[:4], 0644); err != nil {
		t.Fatal(err)
	}
	m := &manager{models: root, progress: map[string]*progress{}}
	if err := m.fetchFile(spec, spec.Files[0], &progress{}, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, spec.ID, "test.gguf")); err != nil {
		t.Fatal(fmt.Errorf("final file: %w", err))
	}
}

func TestRemainingCountsOnlyMissingBytesAndDropsOversizedPartial(t *testing.T) {
	root := t.TempDir()
	spec := modelSpec{ID: "test", Files: []fileSpec{{Name: "one", Size: 10}, {Name: "two", Size: 20}}}
	dir := filepath.Join(root, spec.ID)
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "one.partial"), make([]byte, 4), 0644)
	_ = os.WriteFile(filepath.Join(dir, "two.partial"), make([]byte, 21), 0644)
	m := &manager{models: root}
	if got := m.remaining(spec); got != 26 {
		t.Fatalf("faltaban 26 bytes y se calcularon %d", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "two.partial")); !os.IsNotExist(err) {
		t.Fatal("el parcial mayor que el archivo debía eliminarse")
	}
}

func TestCorruptCompletePartialIsRemoved(t *testing.T) {
	content := []byte("GGUFbad")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(content) }))
	defer server.Close()
	root := t.TempDir()
	spec := modelSpec{ID: "test", Files: []fileSpec{{Name: "bad.gguf", URL: server.URL, SHA256: "wrong", Size: int64(len(content))}}}
	_ = os.MkdirAll(filepath.Join(root, spec.ID), 0755)
	m := &manager{models: root, progress: map[string]*progress{}}
	if err := m.fetchFile(spec, spec.Files[0], &progress{}, 0); err == nil {
		t.Fatal("se esperaba fallo de checksum")
	}
	if _, err := os.Stat(filepath.Join(root, spec.ID, "bad.gguf.partial")); !os.IsNotExist(err) {
		t.Fatal("el parcial corrupto debía eliminarse")
	}
}

func TestDownloadStopsAfterFiveRedirects(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
		w.Header().Set("Location", server.URL+"/"+strconv.Itoa(n+1))
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	root := t.TempDir()
	spec := modelSpec{ID: "test", Files: []fileSpec{{Name: "x", URL: server.URL + "/0", Size: 4}}}
	_ = os.MkdirAll(filepath.Join(root, spec.ID), 0755)
	m := &manager{models: root, progress: map[string]*progress{}}
	if err := m.fetchFile(spec, spec.Files[0], &progress{}, 0); err == nil || !strings.Contains(err.Error(), "too many redirects") {
		t.Fatalf("se esperaba límite de redirecciones y llegó %v", err)
	}
}

func TestEffectiveMemoryUsesTheSmallestCgroupLimit(t *testing.T) {
	files := map[string]string{"/sys/fs/cgroup/memory.max": "1000", "/sys/fs/cgroup/memory.current": "400"}
	read := func(path string) ([]byte, error) {
		if value, ok := files[path]; ok {
			return []byte(value), nil
		}
		return nil, os.ErrNotExist
	}
	if got := effectiveMemory(2000, read); got != 600 {
		t.Fatalf("se esperaba 600 bytes efectivos y llegaron %d", got)
	}
}

func TestSupervisorRestartsAndIncreasesBackoff(t *testing.T) {
	root := t.TempDir()
	m := newManager(root, "token")
	var starts atomic.Int32
	m.command = func(modelSpec) *exec.Cmd { starts.Add(1); return exec.Command("sh", "-c", "exit 1") }
	m.health = func(context.Context) error { return nil }
	delaySeen := make(chan time.Duration, 2)
	var waits atomic.Int32
	m.wait = func(_ context.Context, delay time.Duration) bool {
		delaySeen <- delay
		return waits.Add(1) == 1
	}
	spec := modelSpec{ID: "test", Files: []fileSpec{{Name: "x"}}}
	if err := m.activateModel(spec); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []time.Duration{time.Second, 2 * time.Second} {
		select {
		case delay := <-delaySeen:
			if delay != expected {
				t.Fatalf("se esperaba %s y llegó %s", expected, delay)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("el supervisor no observó la caída")
		}
	}
	if starts.Load() != 2 {
		t.Fatalf("se esperaban dos procesos y se iniciaron %d", starts.Load())
	}
	m.stopProcess()
}

func TestRestartDelayGrowsAndStopsAtSixtySeconds(t *testing.T) {
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, 60 * time.Second, 60 * time.Second}
	for i, expected := range want {
		if got := restartDelay(i); got != expected {
			t.Fatalf("intento %d: %s", i, got)
		}
	}
}

func TestSwitchUsesCurrentRSSAndMeasuresAgain(t *testing.T) {
	m := newManager(t.TempDir(), "token")
	m.command = func(modelSpec) *exec.Cmd { return exec.Command("sh", "-c", "sleep 60") }
	m.health = func(context.Context) error { return nil }
	previous := catalog[0]
	if err := m.activateModel(previous); err != nil {
		t.Fatal(err)
	}
	var reads atomic.Int32
	m.available = func() int64 {
		if reads.Add(1) == 1 {
			return 3 << 30
		}
		return 5 << 30
	}
	m.rss = func(*exec.Cmd) int64 { return 2 << 30 }
	target := modelSpec{ID: "test-3b", RAM: 4 << 30}
	if err := m.switchModel(target); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	active, healthy := m.active, m.healthy
	m.mu.Unlock()
	if active != target.ID || !healthy {
		t.Fatalf("activo=%q sano=%v", active, healthy)
	}
	if reads.Load() != 2 {
		t.Fatalf("se esperaban dos mediciones y hubo %d", reads.Load())
	}
	m.stopProcess()
}

func TestImpossibleSwitchKeepsCurrentModel(t *testing.T) {
	m := newManager(t.TempDir(), "token")
	m.command = func(modelSpec) *exec.Cmd { return exec.Command("sh", "-c", "sleep 60") }
	m.health = func(context.Context) error { return nil }
	previous := catalog[0]
	if err := m.activateModel(previous); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	pid := m.process.Process.Pid
	m.mu.Unlock()
	m.available = func() int64 { return 3 << 30 }
	m.rss = func(*exec.Cmd) int64 { return 2 << 30 }
	err := m.switchModel(modelSpec{ID: "test-7b", RAM: 6 << 30})
	var memory *memoryError
	if !errors.As(err, &memory) {
		t.Fatalf("se esperaba error de memoria y llegó %v", err)
	}
	m.mu.Lock()
	active, healthy, currentPID := m.active, m.healthy, m.process.Process.Pid
	m.mu.Unlock()
	if active != previous.ID || !healthy || currentPID != pid {
		t.Fatalf("el modelo anterior cambió: activo=%q sano=%v pid=%d/%d", active, healthy, currentPID, pid)
	}
	m.stopProcess()
}

func TestFailedSwitchRestoresHealthyPreviousModel(t *testing.T) {
	m := newManager(t.TempDir(), "token")
	m.command = func(modelSpec) *exec.Cmd { return exec.Command("sh", "-c", "sleep 60") }
	var healthCalls atomic.Int32
	m.health = func(context.Context) error {
		if healthCalls.Add(1) == 2 {
			return errors.New("target unhealthy")
		}
		return nil
	}
	previous := catalog[0]
	if err := m.activateModel(previous); err != nil {
		t.Fatal(err)
	}
	m.available = func() int64 { return 5 << 30 }
	m.rss = func(*exec.Cmd) int64 { return 2 << 30 }
	err := m.switchModel(modelSpec{ID: "test-3b", RAM: 4 << 30})
	if err == nil || !strings.Contains(err.Error(), "target unhealthy") {
		t.Fatalf("error=%v", err)
	}
	m.mu.Lock()
	active, healthy := m.active, m.healthy
	m.mu.Unlock()
	if active != previous.ID || !healthy {
		t.Fatalf("no se restauró: activo=%q sano=%v", active, healthy)
	}
	m.stopProcess()
}
