package check

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// statusServer démarre un serveur local qui répond toujours `code`.
func statusServer(t *testing.T, code int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code >= 300 && code < 400 {
			w.Header().Set("Location", "/ailleurs")
		}
		w.WriteHeader(code)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		name   string
		code   int
		want   int
		wantOK bool
	}{
		{"200 accepté par défaut", 200, 0, true},
		{"204 accepté par défaut", 204, 0, true},
		{"301 accepté par défaut, redirection non suivie", 301, 0, true},
		{"404 refusé par défaut", 404, 0, false},
		{"500 refusé par défaut", 500, 0, false},
		{"--status 200 et réponse 200", 200, 200, true},
		{"--status 200 mais réponse 201", 201, 200, false},
		{"--status 404 et réponse 404", 404, 404, true},
		{"--status 301 vérifiable car redirection non suivie", 301, 301, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url := statusServer(t, c.code)
			res := Target(url, Options{WantStatus: c.want, Timeout: 2 * time.Second})
			if res.OK != c.wantOK {
				t.Errorf("OK = %v, want %v (detail: %s)", res.OK, c.wantOK, res.Detail)
			}
			if !strings.Contains(res.Detail, "HTTP") {
				t.Errorf("Detail = %q, want le code HTTP", res.Detail)
			}
			if res.Target != url {
				t.Errorf("Target = %q, want %q", res.Target, url)
			}
		})
	}
}

func TestHTTPWrongStatusMentionsExpected(t *testing.T) {
	res := Target(statusServer(t, 503), Options{WantStatus: 200})
	if res.OK || !strings.Contains(res.Detail, "attendu 200") {
		t.Errorf("Result = %+v, want KO mentionnant le code attendu", res)
	}
}

func TestHTTPTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(release) })

	start := time.Now()
	res := Target(srv.URL, Options{Timeout: 200 * time.Millisecond})
	if res.OK {
		t.Fatal("OK = true, want false quand le serveur ne répond pas à temps")
	}
	if res.Detail != "délai dépassé" {
		t.Errorf("Detail = %q, want \"délai dépassé\"", res.Detail)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("Target() a mis %s, want ~200ms", elapsed)
	}
}

func TestHTTPServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	if res := Target(url, Options{Timeout: 3 * time.Second}); res.OK {
		t.Errorf("OK = true, want false sur un serveur arrêté")
	}
}

func TestTCPOpenPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	res := Target(ln.Addr().String(), Options{Timeout: 2 * time.Second})
	if !res.OK || res.Detail != "port ouvert" {
		t.Errorf("Result = %+v, want OK \"port ouvert\"", res)
	}
}

func TestTCPClosedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // le port est maintenant fermé

	if res := Target(addr, Options{Timeout: 3 * time.Second}); res.OK {
		t.Errorf("OK = true, want false sur un port fermé (%s)", addr)
	}
}

func TestInvalidTarget(t *testing.T) {
	for _, target := range []string{"localhost", "ftp://example.com", "", "localhost:0", "localhost:99999", "localhost:http"} {
		t.Run(target, func(t *testing.T) {
			res := Target(target, Options{})
			if res.OK || !strings.Contains(res.Detail, "cible invalide") {
				t.Errorf("Result = %+v, want KO \"cible invalide\"", res)
			}
		})
	}
}

func TestZeroOptions(t *testing.T) {
	// Options vides : délai par défaut et "tout code < 400" accepté.
	res := Target(statusServer(t, 200), Options{})
	if !res.OK {
		t.Errorf("Result = %+v, want OK avec le délai par défaut", res)
	}
}
