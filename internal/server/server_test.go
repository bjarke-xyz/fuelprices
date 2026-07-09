package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testDist() fstest.MapFS {
	return fstest.MapFS{
		"index.html":          {Data: []byte(`<div id="root"></div>`)},
		"assets/index-abc.js": {Data: []byte(`console.log(1)`)},
	}
}

func TestSPARouting(t *testing.T) {
	srv := httptest.NewServer(spaHandler(testDist()))
	defer srv.Close()

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", http.StatusOK, `<div id="root"></div>`},
		// Client-side routes fall back to the shell.
		{"/historical", http.StatusOK, `<div id="root"></div>`},
		{"/overunder", http.StatusOK, `<div id="root"></div>`},
		{"/assets/index-abc.js", http.StatusOK, `console.log(1)`},
		// A stale or missing asset must 404 rather than serve HTML.
		{"/assets/index-gone.js", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := srv.Client().Get(srv.URL + tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantBody != "" {
				body := make([]byte, len(tt.wantBody))
				resp.Body.Read(body)
				if string(body) != tt.wantBody {
					t.Errorf("body = %q, want %q", body, tt.wantBody)
				}
			}
		})
	}
}

func TestJobRequiresKey(t *testing.T) {
	handler := New(nil, nil, "secret", testDist()).Handler()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"wrong key", "nope", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/job", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// An unset JOB_KEY must not turn the job endpoint into an open door.
func TestJobRejectsEmptyKey(t *testing.T) {
	handler := New(nil, nil, "", testDist()).Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/job", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
