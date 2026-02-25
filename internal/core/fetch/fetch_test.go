package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/evilCYH/NodeHub/internal/core/mihomo"
)

func TestParseSubscriptionUserInfo(t *testing.T) {
	tests := []struct {
		name   string
		header string
		check  func(t *testing.T, info *SubInfo)
	}{
		{
			name:   "full header",
			header: "upload=1; download=2; total=3; expire=4",
			check: func(t *testing.T, info *SubInfo) {
				if info == nil {
					t.Fatalf("expected info, got nil")
				}
				if info.Upload != 1 || !info.HasUpload() {
					t.Fatalf("unexpected upload: %v", info.Upload)
				}
				if info.Download != 2 || !info.HasDownload() {
					t.Fatalf("unexpected download: %v", info.Download)
				}
				if info.Total != 3 || !info.HasTotal() {
					t.Fatalf("unexpected total: %v", info.Total)
				}
				if info.Expire != 4 || !info.HasExpire() {
					t.Fatalf("unexpected expire: %v", info.Expire)
				}
			},
		},
		{
			name:   "partial fields and unknown key",
			header: " upload=10 ; foo=bar ; total=20 ",
			check: func(t *testing.T, info *SubInfo) {
				if info == nil {
					t.Fatalf("expected info, got nil")
				}
				if !info.HasUpload() || info.Upload != 10 {
					t.Fatalf("expected upload=10, got %v", info.Upload)
				}
				if !info.HasTotal() || info.Total != 20 {
					t.Fatalf("expected total=20, got %v", info.Total)
				}
				if info.HasDownload() {
					t.Fatalf("download should be unset")
				}
				if info.HasExpire() {
					t.Fatalf("expire should be unset")
				}
			},
		},
		{
			name:   "invalid number ignored",
			header: "upload=abc; download=2",
			check: func(t *testing.T, info *SubInfo) {
				if info == nil {
					t.Fatalf("expected info, got nil")
				}
				if info.HasUpload() {
					t.Fatalf("upload should be unset")
				}
				if !info.HasDownload() || info.Download != 2 {
					t.Fatalf("expected download=2, got %v", info.Download)
				}
			},
		},
		{
			name:   "all zero values are valid",
			header: "upload=0; download=0; total=0; expire=0",
			check: func(t *testing.T, info *SubInfo) {
				if info == nil {
					t.Fatalf("expected info, got nil")
				}
				if !info.HasUpload() || !info.HasDownload() || !info.HasTotal() || !info.HasExpire() {
					t.Fatalf("all fields should be marked valid")
				}
				if info.Upload != 0 || info.Download != 0 || info.Total != 0 || info.Expire != 0 {
					t.Fatalf("expected all values to be 0")
				}
			},
		},
		{
			name:   "empty header returns nil",
			header: "",
			check: func(t *testing.T, info *SubInfo) {
				if info != nil {
					t.Fatalf("expected nil info, got %+v", info)
				}
			},
		},
		{
			name:   "all invalid fields returns nil",
			header: "foo=bar; upload=abc",
			check: func(t *testing.T, info *SubInfo) {
				if info != nil {
					t.Fatalf("expected nil info, got %+v", info)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t, parseSubscriptionUserInfo(tc.header))
		})
	}
}

type mockFetchResponse struct {
	status int
	body   string
	info   string
}

func withFetchRounds(t *testing.T, rounds [][]string) {
	t.Helper()
	old := fetchRoundUserAgents
	fetchRoundUserAgents = rounds
	t.Cleanup(func() {
		fetchRoundUserAgents = old
	})
}

func newTestHTTPClient() *mihomo.HC {
	return &mihomo.HC{
		Client: &http.Client{},
	}
}

func TestFetchSubscriptionWithFallback_ContentFromUA1InfoFromUA2(t *testing.T) {
	withFetchRounds(t, [][]string{{"Clash", "clash-verge/v2.4.0", "v2rayNG/1.8.12"}})

	responses := map[string]mockFetchResponse{
		"Clash": {
			status: http.StatusOK,
			body:   "content-from-ua1",
		},
		"clash-verge/v2.4.0": {
			status: http.StatusOK,
			body:   "content-from-ua2-ignored",
			info:   "upload=1; download=2; total=3; expire=4",
		},
	}

	var mu sync.Mutex
	requestedUA := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		mu.Lock()
		requestedUA = append(requestedUA, ua)
		mu.Unlock()

		resp, ok := responses[ua]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if resp.info != "" {
			w.Header().Set("subscription-userinfo", resp.info)
		}
		w.WriteHeader(resp.status)
		_, _ = w.Write([]byte(resp.body))
	}))
	defer server.Close()

	content, info, err := fetchSubscriptionWithFallback(context.Background(), newTestHTTPClient(), server.URL, func(_, _, _ string) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(content) != "content-from-ua1" {
		t.Fatalf("expected content from ua1, got %q", string(content))
	}
	if info == nil {
		t.Fatalf("expected info from ua2, got nil")
	}
	if !info.HasUpload() || info.Upload != 1 || !info.HasDownload() || info.Download != 2 || !info.HasTotal() || info.Total != 3 || !info.HasExpire() || info.Expire != 4 {
		t.Fatalf("unexpected info: %+v", info)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requestedUA) != 2 {
		t.Fatalf("expected two requests, got %d (%v)", len(requestedUA), requestedUA)
	}
	if requestedUA[0] != "Clash" || requestedUA[1] != "clash-verge/v2.4.0" {
		t.Fatalf("unexpected request order: %v", requestedUA)
	}
}

func TestFetchSubscriptionWithFallback_ReturnsUA1ContentWhenUA2UA3Fail(t *testing.T) {
	withFetchRounds(t, [][]string{{"Clash", "clash-verge/v2.4.0", "v2rayNG/1.8.12"}})

	responses := map[string]mockFetchResponse{
		"Clash": {
			status: http.StatusOK,
			body:   "content-from-ua1",
		},
		"clash-verge/v2.4.0": {
			status: http.StatusInternalServerError,
		},
		"v2rayNG/1.8.12": {
			status: http.StatusForbidden,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		resp, ok := responses[ua]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(resp.status)
		_, _ = w.Write([]byte(resp.body))
	}))
	defer server.Close()

	content, info, err := fetchSubscriptionWithFallback(context.Background(), newTestHTTPClient(), server.URL, func(_, _, _ string) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(content) != "content-from-ua1" {
		t.Fatalf("expected content from ua1, got %q", string(content))
	}
	if info != nil {
		t.Fatalf("expected nil info, got %+v", info)
	}
}

func TestFetchSubscriptionWithFallback_FailsOnNon2xx(t *testing.T) {
	withFetchRounds(t, [][]string{{"Clash"}})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	content, info, err := fetchSubscriptionWithFallback(context.Background(), newTestHTTPClient(), server.URL, func(_, _, _ string) {})
	if err == nil {
		t.Fatalf("expected error for non-2xx response")
	}
	if content != nil {
		t.Fatalf("expected nil content, got %q", string(content))
	}
	if info != nil {
		t.Fatalf("expected nil info, got %+v", info)
	}
}
