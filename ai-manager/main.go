package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type fileSpec struct {
	Name, URL, SHA256 string
	Size              int64
}
type modelSpec struct {
	ID, Name, Version, License, Source, Alias string
	RAM                                       int64
	Files                                     []fileSpec
}
type progress struct {
	Downloaded int64
	Running    bool
	Err        string
}
type modelView struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	License         string `json:"license"`
	Source          string `json:"source"`
	Checksum        string `json:"checksum"`
	DownloadBytes   int64  `json:"downloadBytes"`
	RAMBytes        int64  `json:"ramBytes"`
	Installed       bool   `json:"installed"`
	Active          bool   `json:"active"`
	Healthy         bool   `json:"healthy"`
	Downloading     bool   `json:"downloading"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	Error           string `json:"error,omitempty"`
}

var catalog = []modelSpec{
	{ID: "qwen2.5-1.5b-instruct", Name: "Qwen2.5-1.5B-Instruct Q4_K_M", Version: "main", License: "Apache-2.0", Source: "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF", Alias: "qwen2.5-1.5b-instruct", RAM: 2 << 30, Files: []fileSpec{{"qwen2.5-1.5b-instruct-q4_k_m.gguf", "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf", "6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e", 1117320736}}},
	{ID: "qwen2.5-3b-instruct", Name: "Qwen2.5-3B-Instruct Q4_K_M", Version: "main", License: "Qwen Research", Source: "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF", Alias: "qwen2.5-3b-instruct", RAM: 4 << 30, Files: []fileSpec{{"qwen2.5-3b-instruct-q4_k_m.gguf", "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf", "626b4a6678b86442240e33df819e00132d3ba7dddfe1cdc4fbb18e0a9615c62d", 2104932768}}},
	{ID: "qwen2.5-7b-instruct", Name: "Qwen2.5-7B-Instruct Q4_K_M", Version: "main", License: "Apache-2.0", Source: "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF", Alias: "qwen2.5-7b-instruct", RAM: 6 << 30, Files: []fileSpec{
		{"qwen2.5-7b-instruct-q4_k_m-00001-of-00002.gguf", "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2.5-7b-instruct-q4_k_m-00001-of-00002.gguf", "dfce12e3862a5283ccfb88221b48480e58745165de856439950d0f22590580db", 3993201344},
		{"qwen2.5-7b-instruct-q4_k_m-00002-of-00002.gguf", "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2.5-7b-instruct-q4_k_m-00002-of-00002.gguf", "539cf93f78e887edea1c04e2d7d8cdaca9d01dae9c9025bcb8accbe29df3d72a", 689872288},
	}},
}

type manager struct {
	mu            sync.Mutex
	activationMu  sync.Mutex
	progress      map[string]*progress
	active        string
	process       *exec.Cmd
	cancelProcess context.CancelFunc
	processDone   chan struct{}
	healthy       bool
	lastError     string
	models, token string
	command       func(modelSpec) *exec.Cmd
	health        func(context.Context) error
	wait          func(context.Context, time.Duration) bool
	available     func() int64
	rss           func(*exec.Cmd) int64
}

type memoryError struct {
	Required  int64 `json:"requiredBytes"`
	Available int64 `json:"availableBytes"`
}

func (e *memoryError) Error() string { return "insufficient memory" }

func main() {
	m := newManager(env("MODELS_PATH", "/models"), os.Getenv("AI_MANAGER_TOKEN"))
	if m.token == "" {
		panic("AI_MANAGER_TOKEN is required")
	}
	_ = os.MkdirAll(m.models, 0755)
	if data, err := os.ReadFile(filepath.Join(m.models, "active")); err == nil {
		m.active = strings.TrimSpace(string(data))
		if s := find(m.active); s != nil && m.installed(*s) {
			go func() { _ = m.activateModel(*s) }()
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /models", m.auth(m.list))
	mux.HandleFunc("POST /models/{id}/download", m.auth(m.download))
	mux.HandleFunc("POST /models/{id}/activate", m.auth(m.activate))
	mux.HandleFunc("DELETE /models/{id}", m.auth(m.remove))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	if err := http.ListenAndServe(":8081", mux); err != nil {
		panic(err)
	}
}

func newManager(models, token string) *manager {
	m := &manager{progress: map[string]*progress{}, models: models, token: token}
	m.command = func(s modelSpec) *exec.Cmd {
		args := []string{"--model", filepath.Join(m.dir(s), s.Files[0].Name), "--alias", s.Alias, "--host", "0.0.0.0", "--port", "8080", "--ctx-size", "4096", "--threads", "2", "--parallel", "1", "--n-predict", "512", "--no-webui"}
		return exec.Command("/app/llama-server", args...)
	}
	m.health = modelHealth
	m.wait = waitContext
	m.available = availableMemory
	m.rss = processRSS
	return m
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func find(id string) *modelSpec {
	for i := range catalog {
		if catalog[i].ID == id {
			return &catalog[i]
		}
	}
	return nil
}
func (m *manager) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+m.token {
			http.Error(w, "unauthorized", 401)
			return
		}
		next(w, r)
	}
}
func (m *manager) dir(s modelSpec) string { return filepath.Join(m.models, s.ID) }
func (m *manager) installed(s modelSpec) bool {
	for _, f := range s.Files {
		if st, e := os.Stat(filepath.Join(m.dir(s), f.Name)); e != nil || st.Size() != f.Size {
			return false
		}
	}
	return true
}
func (m *manager) list(w http.ResponseWriter, _ *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	views := make([]modelView, 0, len(catalog))
	for _, s := range catalog {
		p := m.progress[s.ID]
		v := modelView{ID: s.ID, Name: s.Name, Version: s.Version, License: s.License, Source: s.Source, Checksum: checksums(s), DownloadBytes: total(s), RAMBytes: s.RAM, Installed: m.installed(s), Active: m.active == s.ID, Healthy: m.active == s.ID && m.healthy}
		if v.Active && m.lastError != "" {
			v.Error = m.lastError
		}
		if p != nil {
			v.Downloading = p.Running
			v.DownloadedBytes = p.Downloaded
			v.Error = p.Err
		}
		views = append(views, v)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"models": views, "diskAvailableBytes": diskAvailable(m.models)})
}
func checksums(s modelSpec) string {
	v := []string{}
	for _, f := range s.Files {
		v = append(v, f.SHA256)
	}
	return strings.Join(v, " + ")
}
func total(s modelSpec) (n int64) {
	for _, f := range s.Files {
		n += f.Size
	}
	return
}
func (m *manager) remaining(s modelSpec) (n int64) {
	for _, f := range s.Files {
		if st, err := os.Stat(filepath.Join(m.dir(s), f.Name)); err == nil && st.Size() == f.Size {
			continue
		}
		partial := filepath.Join(m.dir(s), f.Name+".partial")
		var have int64
		if st, err := os.Stat(partial); err == nil {
			have = st.Size()
			if have > f.Size {
				_ = os.Remove(partial)
				have = 0
			}
		}
		n += f.Size - have
	}
	return n
}
func diskAvailable(path string) int64 {
	var st syscall.Statfs_t
	if syscall.Statfs(path, &st) != nil {
		return 0
	}
	return int64(st.Bavail) * int64(st.Bsize)
}
func (m *manager) download(w http.ResponseWriter, r *http.Request) {
	s := find(r.PathValue("id"))
	if s == nil {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("acceptLicense") != "true" {
		http.Error(w, "license acceptance required", 422)
		return
	}
	m.mu.Lock()
	if p := m.progress[s.ID]; p != nil && p.Running {
		m.mu.Unlock()
		w.WriteHeader(202)
		return
	}
	if diskAvailable(m.models) < m.remaining(*s)+(256<<20) {
		m.mu.Unlock()
		http.Error(w, "insufficient disk", 422)
		return
	}
	p := &progress{Running: true}
	m.progress[s.ID] = p
	m.mu.Unlock()
	go m.fetch(*s, p)
	w.WriteHeader(202)
}
func (m *manager) fetch(s modelSpec, p *progress) {
	err := os.MkdirAll(m.dir(s), 0755)
	var completed int64
	for _, f := range s.Files {
		if err == nil {
			err = m.fetchFile(s, f, p, completed)
			if err == nil {
				completed += f.Size
			}
		}
	}
	m.mu.Lock()
	p.Running = false
	if err != nil {
		p.Err = err.Error()
	}
	m.mu.Unlock()
}
func (m *manager) fetchFile(s modelSpec, f fileSpec, p *progress, completed int64) error {
	dest := filepath.Join(m.dir(s), f.Name)
	partial := dest + ".partial"
	var offset int64
	if st, e := os.Stat(partial); e == nil {
		offset = st.Size()
	}
	req, _ := http.NewRequest(http.MethodGet, f.URL, nil)
	if offset > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")
	}
	resp, e := downloadClient().Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return fmt.Errorf("download status %d", resp.StatusCode)
	}
	flags := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == 206 {
		flags |= os.O_APPEND
	} else {
		offset = 0
		flags |= os.O_TRUNC
	}
	out, e := os.OpenFile(partial, flags, 0644)
	if e != nil {
		return e
	}
	counter := &countWriter{m: m, p: p, offset: completed + offset}
	_, e = io.Copy(out, io.TeeReader(io.LimitReader(resp.Body, f.Size-offset+1), counter))
	closeErr := out.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	if e = validate(partial, f); e != nil {
		// Un archivo completo que no pasa formato o checksum no sirve para reanudar: conservarlo
		// convertiría cada nuevo intento en el mismo fallo.
		_ = os.Remove(partial)
		return e
	}
	return os.Rename(partial, dest)
}

func downloadClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{ResponseHeaderTimeout: 30 * time.Second},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("invalid redirect scheme")
			}
			return nil
		},
	}
}

type countWriter struct {
	m         *manager
	p         *progress
	offset, n int64
}

func (c *countWriter) Write(b []byte) (int, error) {
	c.n += int64(len(b))
	c.m.mu.Lock()
	c.p.Downloaded = c.offset + c.n
	c.m.mu.Unlock()
	return len(b), nil
}
func validate(path string, s fileSpec) error {
	st, e := os.Stat(path)
	if e != nil || st.Size() != s.Size {
		return errors.New("invalid size")
	}
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, e = io.ReadFull(f, magic); e != nil || string(magic) != "GGUF" {
		return errors.New("invalid GGUF")
	}
	_, _ = f.Seek(0, 0)
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return e
	}
	if hex.EncodeToString(h.Sum(nil)) != s.SHA256 {
		return errors.New("invalid checksum")
	}
	return nil
}
func (m *manager) activate(w http.ResponseWriter, r *http.Request) {
	s := find(r.PathValue("id"))
	if s == nil {
		http.NotFound(w, r)
		return
	}
	if !m.installed(*s) {
		http.Error(w, "not installed", 422)
		return
	}
	if err := m.switchModel(*s); err != nil {
		var memory *memoryError
		if errors.As(err, &memory) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"key": "settings.ai.insufficientMemory", "requiredBytes": memory.Required,
				"availableBytes": memory.Available,
			})
			return
		}
		http.Error(w, err.Error(), 502)
		return
	}
	w.WriteHeader(204)
}

func processRSS(cmd *exec.Cmd) int64 {
	if cmd == nil || cmd.Process == nil {
		return 0
	}
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(cmd.Process.Pid), "status"))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0
			}
			value, _ := strconv.ParseInt(fields[1], 10, 64)
			return value * 1024
		}
	}
	return 0
}
func availableMemory() int64 {
	data, _ := os.ReadFile("/proc/meminfo")
	host := int64(0)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemAvailable:") {
			f := strings.Fields(line)
			n, _ := strconv.ParseInt(f[1], 10, 64)
			host = n * 1024
			break
		}
	}
	return effectiveMemory(host, os.ReadFile)
}

func effectiveMemory(host int64, read func(string) ([]byte, error)) int64 {
	remaining := func(limitPath, usedPath string) int64 {
		limitRaw, err := read(limitPath)
		if err != nil || strings.TrimSpace(string(limitRaw)) == "max" {
			return 0
		}
		usedRaw, err := read(usedPath)
		if err != nil {
			return 0
		}
		limit, err1 := strconv.ParseInt(strings.TrimSpace(string(limitRaw)), 10, 64)
		used, err2 := strconv.ParseInt(strings.TrimSpace(string(usedRaw)), 10, 64)
		if err1 != nil || err2 != nil || limit <= used || limit > 1<<60 {
			return 0
		}
		return limit - used
	}
	container := remaining("/sys/fs/cgroup/memory.max", "/sys/fs/cgroup/memory.current")
	if container == 0 {
		container = remaining("/sys/fs/cgroup/memory/memory.limit_in_bytes", "/sys/fs/cgroup/memory/memory.usage_in_bytes")
	}
	if container > 0 && (host == 0 || container < host) {
		return container
	}
	return host
}
func (m *manager) activateModel(s modelSpec) error {
	m.stopProcess()
	ctx, cancel := context.WithCancel(context.Background())
	cmd, err := m.launch(ctx, s)
	if err != nil {
		cancel()
		return err
	}
	done := make(chan struct{})
	m.mu.Lock()
	m.active = s.ID
	m.process = cmd
	m.cancelProcess = cancel
	m.processDone = done
	m.healthy = true
	m.lastError = ""
	_ = os.WriteFile(filepath.Join(m.models, "active"), []byte(s.ID), 0644)
	m.mu.Unlock()
	go m.supervise(ctx, s, cmd, done)
	return nil
}

func (m *manager) switchModel(target modelSpec) error {
	m.activationMu.Lock()
	defer m.activationMu.Unlock()

	m.mu.Lock()
	previousID, previousProcess, previousHealthy := m.active, m.process, m.healthy
	m.mu.Unlock()
	if previousID == target.ID && previousHealthy {
		return nil
	}
	available := m.available()
	available += m.rss(previousProcess)
	if available < target.RAM {
		return &memoryError{Required: target.RAM, Available: available}
	}

	var previous *modelSpec
	if previousID != "" && previousID != target.ID {
		previous = find(previousID)
	}
	m.stopProcess()
	available = m.available()
	if available < target.RAM {
		err := &memoryError{Required: target.RAM, Available: available}
		return m.restore(previous, err)
	}
	if err := m.activateModel(target); err != nil {
		return m.restore(previous, err)
	}
	return nil
}

func (m *manager) restore(previous *modelSpec, cause error) error {
	if previous == nil {
		return cause
	}
	if err := m.activateModel(*previous); err != nil {
		return fmt.Errorf("%w; previous model could not be restored: %v", cause, err)
	}
	return cause
}

func (m *manager) launch(ctx context.Context, s modelSpec) (*exec.Cmd, error) {
	cmd := m.command(s)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := m.health(healthCtx); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, err
	}
	return cmd, nil
}

func modelHealth(ctx context.Context) error {
	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:8080/health", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		if !waitContext(ctx, time.Second) {
			return errors.New("model did not become healthy")
		}
	}
}

func waitContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func restartDelay(attempt int) time.Duration {
	if attempt > 6 {
		attempt = 6
	}
	d := time.Second << attempt
	if d > 60*time.Second {
		return 60 * time.Second
	}
	return d
}

func (m *manager) supervise(ctx context.Context, s modelSpec, cmd *exec.Cmd, done chan struct{}) {
	defer close(done)
	attempt := 0
	for {
		err := cmd.Wait()
		m.mu.Lock()
		if m.process == cmd {
			m.process = nil
			m.healthy = false
		}
		if err != nil {
			m.lastError = err.Error()
		}
		m.mu.Unlock()
		if ctx.Err() != nil {
			return
		}
		if !m.wait(ctx, restartDelay(attempt)) {
			return
		}
		attempt++
		next, launchErr := m.launch(ctx, s)
		if launchErr != nil {
			m.mu.Lock()
			m.lastError = launchErr.Error()
			m.mu.Unlock()
			continue
		}
		cmd = next
		m.mu.Lock()
		m.process = cmd
		m.healthy = true
		m.lastError = ""
		m.mu.Unlock()
	}
}

func (m *manager) stopProcess() {
	m.mu.Lock()
	cancel, cmd, done := m.cancelProcess, m.process, m.processDone
	m.cancelProcess, m.process, m.processDone, m.healthy = nil, nil, nil, false
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	if done != nil {
		<-done
	}
}
func (m *manager) remove(w http.ResponseWriter, r *http.Request) {
	s := find(r.PathValue("id"))
	if s == nil {
		http.NotFound(w, r)
		return
	}
	m.mu.Lock()
	active := m.active == s.ID
	m.mu.Unlock()
	if active {
		http.Error(w, "active model cannot be removed", 409)
		return
	}
	if err := os.RemoveAll(m.dir(*s)); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
