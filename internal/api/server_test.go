package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"safora/internal/database"
)

type harness struct {
	t      *testing.T
	srv    *Server
	ts     *httptest.Server
	token  string
	tmp    string
	client *http.Client
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	tmp := t.TempDir()
	db, err := database.InitDB(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	srv, err := NewServer(db, filepath.Join(tmp, "test.token"))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewUnstartedServer(nil)
	ts.Config.Handler = srv.Handler(ts.Listener.Addr().String())
	ts.Start()
	t.Cleanup(ts.Close)

	return &harness{t: t, srv: srv, ts: ts, token: srv.token, tmp: tmp, client: ts.Client()}
}

// call sends a request with the API token. mod can adjust it (headers, Host).
func (h *harness) call(method, path string, body any, mod ...func(*http.Request)) (*http.Response, string) {
	h.t.Helper()
	var rd io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rd = strings.NewReader(b)
	case []byte:
		rd = bytes.NewReader(b)
	default:
		raw, _ := json.Marshal(b)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, h.ts.URL+path, rd)
	if err != nil {
		h.t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+h.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, m := range mod {
		m(req)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		h.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp, string(out)
}

func (h *harness) status(method, path string, body any, mod ...func(*http.Request)) int {
	h.t.Helper()
	resp, _ := h.call(method, path, body, mod...)
	return resp.StatusCode
}

func job(name string, srcs, dsts []string, extra map[string]any) map[string]any {
	j := map[string]any{"Name": name, "StorageStrategy": "Date-Stamped Mirroring"}
	var s, d []map[string]string
	for _, p := range srcs {
		s = append(s, map[string]string{"Path": p})
	}
	for _, p := range dsts {
		d = append(d, map[string]string{"Path": p})
	}
	j["Sources"], j["Destinations"] = s, d
	for k, v := range extra {
		j[k] = v
	}
	return j
}

func (h *harness) create(body map[string]any) int64 {
	h.t.Helper()
	resp, out := h.call("POST", "/api/jobs", body)
	if resp.StatusCode != 200 {
		h.t.Fatalf("create job: %d %s", resp.StatusCode, out)
	}
	var created struct{ ID int64 }
	json.Unmarshal([]byte(out), &created)
	return created.ID
}

func TestAuthentication(t *testing.T) {
	h := newHarness(t)

	none := func(r *http.Request) { r.Header.Del("Authorization") }
	wrong := func(r *http.Request) { r.Header.Set("Authorization", "Bearer nope") }
	if got := h.status("GET", "/api/jobs", nil, none); got != 401 {
		t.Errorf("no token: %d, want 401", got)
	}
	if got := h.status("GET", "/api/jobs", nil, wrong); got != 401 {
		t.Errorf("wrong token: %d, want 401", got)
	}
	if got := h.status("GET", "/api/jobs", nil); got != 200 {
		t.Errorf("valid token: %d, want 200", got)
	}
	if got := h.status("GET", "/api/jobs", nil, none, func(r *http.Request) { r.Header.Set("Authorization", "secret-token") }); got != 401 {
		t.Errorf("the old hardcoded credential must not work: %d", got)
	}
}

func TestDashboardReceivesTheTokenAsACookie(t *testing.T) {
	h := newHarness(t)

	resp, _ := h.call("GET", "/", nil, func(r *http.Request) { r.Header.Del("Authorization") })
	if resp.StatusCode != 200 {
		t.Fatalf("static page: %d", resp.StatusCode)
	}
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == cookieName {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected an HttpOnly SameSite=Strict token cookie, got %+v", cookie)
	}

	withCookie := func(r *http.Request) { r.Header.Del("Authorization"); r.AddCookie(cookie) }
	if got := h.status("GET", "/api/jobs", nil, withCookie); got != 200 {
		t.Errorf("the cookie must authenticate the dashboard: %d", got)
	}
}

func TestHostAndOriginChecks(t *testing.T) {
	h := newHarness(t)

	if got := h.status("GET", "/api/jobs", nil, func(r *http.Request) { r.Host = "evil.example" }); got != 403 {
		t.Errorf("foreign Host (DNS rebinding): %d, want 403", got)
	}
	if got := h.status("GET", "/api/jobs", nil, func(r *http.Request) { r.Header.Set("Origin", "http://evil.example") }); got != 403 {
		t.Errorf("foreign Origin: %d, want 403", got)
	}
	if got := h.status("GET", "/api/jobs", nil, func(r *http.Request) { r.Header.Set("Origin", "http://"+r.Host) }); got != 200 {
		t.Errorf("same-origin request: %d, want 200", got)
	}
	resp, _ := h.call("GET", "/api/jobs", nil)
	if v := resp.Header.Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("no CORS header expected, got %q", v)
	}
}

func TestJSONResponsesAndEmptyLists(t *testing.T) {
	h := newHarness(t)
	for _, path := range []string{"/api/jobs", "/api/runs"} {
		resp, body := h.call("GET", path, nil)
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("%s content type = %q", path, ct)
		}
		if strings.TrimSpace(body) != "[]" {
			t.Errorf("%s empty list = %q, want []", path, body)
		}
	}
}

func TestErrorStatuses(t *testing.T) {
	h := newHarness(t)
	cases := []struct {
		method, path string
		want         int
	}{
		{"GET", "/api/jobs/999", 404},
		{"GET", "/api/jobs/abc", 400},
		{"GET", "/api/runs/999", 404},
		{"POST", "/api/jobs/999/run", 404},
		{"POST", "/api/jobs/999/cancel", 409},
		{"DELETE", "/api/jobs/abc", 400},
	}
	for _, c := range cases {
		if got := h.status(c.method, c.path, nil); got != c.want {
			t.Errorf("%s %s = %d, want %d", c.method, c.path, got, c.want)
		}
	}
}

func TestJobValidation(t *testing.T) {
	h := newHarness(t)
	src, dst := t.TempDir(), t.TempDir()

	bad := map[string]map[string]any{
		"no sources":       job("x", nil, []string{dst}, nil),
		"no destinations":  job("x", []string{src}, nil, nil),
		"empty path":       job("x", []string{" "}, []string{dst}, nil),
		"no name":          job("", []string{src}, []string{dst}, nil),
		"invalid schedule": job("x", []string{src}, []string{dst}, map[string]any{"Schedule": "not a cron"}),
	}
	for name, body := range bad {
		if got := h.status("POST", "/api/jobs", body); got != 400 {
			t.Errorf("create with %s: %d, want 400", name, got)
		}
	}

	id := h.create(job("ok", []string{src}, []string{dst}, map[string]any{"Schedule": "0 2 * * *"}))
	if got := h.status("PUT", "/api/jobs/"+itoa(id), job("ok", nil, []string{dst}, nil)); got != 400 {
		t.Errorf("update with no sources: %d, want 400", got)
	}
	if got := h.status("PUT", "/api/jobs/"+itoa(id), job("ok", []string{src}, []string{dst}, map[string]any{"Schedule": "*/x"})); got != 400 {
		t.Errorf("update with invalid schedule: %d, want 400", got)
	}
}

func itoa(n int64) string { b, _ := json.Marshal(n); return string(b) }

func TestJobLifecycleAndListCarriesPaths(t *testing.T) {
	h := newHarness(t)
	src, dst := t.TempDir(), t.TempDir()
	id := h.create(job("nightly", []string{src}, []string{dst}, map[string]any{"Schedule": "0 2 * * *", "SyncDeletions": true}))

	_, body := h.call("GET", "/api/jobs", nil)
	var list []struct {
		Name                  string
		Schedule              string
		SyncDeletions         bool
		Sources, Destinations []struct{ Path string }
	}
	if err := json.Unmarshal([]byte(body), &list); err != nil || len(list) != 1 {
		t.Fatalf("list: %v %s", err, body)
	}
	if list[0].Sources[0].Path != src || list[0].Destinations[0].Path != dst || list[0].Schedule != "0 2 * * *" || !list[0].SyncDeletions {
		t.Errorf("the list must carry sources, destinations and options for the edit form: %+v", list[0])
	}

	if got := h.status("PUT", "/api/jobs/"+itoa(id), job("renamed", []string{src}, []string{dst}, nil)); got != 200 {
		t.Errorf("update: %d", got)
	}
	if got := h.status("DELETE", "/api/jobs/"+itoa(id), nil); got != 204 {
		t.Errorf("delete: %d", got)
	}
	if got := h.status("GET", "/api/jobs/"+itoa(id), nil); got != 404 {
		t.Errorf("deleted job: %d, want 404", got)
	}
}

func TestOnJobsChangedIsCalled(t *testing.T) {
	h := newHarness(t)
	var calls atomic.Int32
	h.srv.OnJobsChanged = func() { calls.Add(1) }

	src, dst := t.TempDir(), t.TempDir()
	id := h.create(job("a", []string{src}, []string{dst}, nil))
	h.call("PUT", "/api/jobs/"+itoa(id), job("b", []string{src}, []string{dst}, nil))
	h.call("DELETE", "/api/jobs/"+itoa(id), nil)
	if calls.Load() != 3 {
		t.Errorf("OnJobsChanged calls = %d, want 3 (create, update, delete)", calls.Load())
	}
}

func TestRequestBodyIsLimited(t *testing.T) {
	h := newHarness(t)
	huge := `{"Name":"` + strings.Repeat("a", 2<<20) + `"}`
	if got := h.status("POST", "/api/jobs", huge); got == 200 {
		t.Error("a 2 MiB body must be rejected")
	}
}

func TestRunConflictAndCancel(t *testing.T) {
	h := newHarness(t)
	src, dst := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644)
	os.MkdirAll(filepath.Join(dst, "a.txt", "blocker"), 0o755) // the copy can never succeed
	id := h.create(job("stuck", []string{src}, []string{dst}, map[string]any{"RetryCount": 5, "RetryWait": 60}))
	path := "/api/jobs/" + itoa(id)

	if got := h.status("POST", path+"/run", nil); got != 202 {
		t.Fatalf("run: %d, want 202", got)
	}
	waitFor(t, "the run to start", func() bool { return h.srv.engine.Busy(id) })

	if got := h.status("POST", path+"/run", nil); got != 409 {
		t.Errorf("second run: %d, want 409", got)
	}
	if got := h.status("POST", path+"/cancel", nil); got != 204 {
		t.Errorf("cancel: %d, want 204", got)
	}
	waitFor(t, "the run to be recorded as cancelled", func() bool {
		_, body := h.call("GET", "/api/runs", nil)
		return strings.Contains(body, `"Status":"cancelled"`)
	})
	if got := h.status("POST", path+"/cancel", nil); got != 409 {
		t.Errorf("cancel with nothing running: %d, want 409", got)
	}
	waitFor(t, "the job to be free", func() bool { return !h.srv.engine.Busy(id) })
	if got := h.status("POST", path+"/run", nil); got != 202 {
		t.Errorf("run after cancel: %d, want 202", got)
	}
	h.srv.engine.Cancel(id)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestFolderListing(t *testing.T) {
	h := newHarness(t)
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "Beta"), 0o755)
	os.MkdirAll(filepath.Join(root, "alpha"), 0o755)
	os.MkdirAll(filepath.Join(root, ".hidden"), 0o755)
	os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o644)

	resp, body := h.call("GET", "/api/fs/list", nil)
	var roots struct{ Entries []struct{ Path string } }
	if resp.StatusCode != 200 || json.Unmarshal([]byte(body), &roots) != nil || len(roots.Entries) == 0 {
		t.Fatalf("roots: %d %s", resp.StatusCode, body)
	}

	resp, body = h.call("GET", "/api/fs/list?path="+root, nil)
	var listing struct {
		Path, Parent string
		Entries      []struct{ Name string }
	}
	if resp.StatusCode != 200 || json.Unmarshal([]byte(body), &listing) != nil {
		t.Fatalf("listing: %d %s", resp.StatusCode, body)
	}
	var names []string
	for _, e := range listing.Entries {
		names = append(names, e.Name)
	}
	want := []string{"alpha", "Beta"}
	if runtime.GOOS == "windows" {
		want = []string{".hidden", "alpha", "Beta"}
	}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("entries = %v, want %v (folders only, sorted, no hidden on Unix)", names, want)
	}
	if listing.Parent != filepath.Dir(root) {
		t.Errorf("parent = %q", listing.Parent)
	}

	checks := map[string]int{
		"relative path": h.status("GET", "/api/fs/list?path=some/dir", nil),
		"missing path":  h.status("GET", "/api/fs/list?path="+filepath.Join(root, "nope"), nil),
		"a file":        h.status("GET", "/api/fs/list?path="+filepath.Join(root, "file.txt"), nil),
	}
	for name, want := range map[string]int{"relative path": 400, "missing path": 404, "a file": 400} {
		if checks[name] != want {
			t.Errorf("%s: %d, want %d", name, checks[name], want)
		}
	}
	if got := h.status("GET", "/api/fs/list?path="+root, nil, func(r *http.Request) { r.Header.Del("Authorization") }); got != 401 {
		t.Errorf("the folder listing must require the token: %d", got)
	}
}

func TestTokenIsGeneratedOncePrivateAndReused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "safora.token")
	first, err := loadOrCreateToken(path)
	if err != nil || len(first) < 32 {
		t.Fatalf("token = %q, err = %v", first, err)
	}
	second, _ := loadOrCreateToken(path)
	if first != second {
		t.Error("the token must persist across restarts")
	}
	if runtime.GOOS != "windows" {
		if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
			t.Errorf("token file mode = %v, want 0600", info.Mode().Perm())
		}
	}
	other, _ := loadOrCreateToken(filepath.Join(t.TempDir(), "other.token"))
	if other == first {
		t.Error("tokens must be random per installation")
	}
}

func TestLiveStreamDeliversBroadcasts(t *testing.T) {
	h := newHarness(t)
	req, _ := http.NewRequest("GET", h.ts.URL+"/api/stream", nil)
	req.Header.Set("Authorization", "Bearer "+h.token)
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type = %q", ct)
	}

	got := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), "data: ") {
				got <- sc.Text()
				return
			}
		}
	}()
	// The subscriber registers just after the headers are sent: broadcast until it hears us.
	deadline := time.After(3 * time.Second)
	for {
		h.srv.broker.Broadcast("[INFO] hello")
		select {
		case line := <-got:
			if line != "data: [INFO] hello" {
				t.Errorf("event = %q", line)
			}
			return
		case <-deadline:
			t.Fatal("no event received")
		case <-time.After(20 * time.Millisecond):
		}
	}
}
