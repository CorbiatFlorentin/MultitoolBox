// Package check fait des smoke tests réseau : une URL répond-elle avec le bon
// code HTTP, un port TCP est-il ouvert ?
package check

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Timeout time.Duration
	// WantStatus est le code HTTP attendu ; 0 accepte tout code < 400.
	// Ignoré pour une cible TCP.
	WantStatus int
}

type Result struct {
	Target   string
	OK       bool
	Detail   string
	Duration time.Duration
}

// Target vérifie une cible : une URL http(s):// est testée par un GET (sans
// suivre les redirections, pour que --status 301 soit vérifiable), tout le
// reste doit être de la forme host:port et est testé par une connexion TCP.
func Target(target string, opts Options) Result {
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	start := time.Now()
	var res Result
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		res = checkHTTP(target, opts)
	} else {
		res = checkTCP(target, opts)
	}
	res.Target = target
	res.Duration = time.Since(start)
	return res
}

func checkHTTP(url string, opts Options) Result {
	client := &http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return Result{Detail: describeErr(err)}
	}
	resp.Body.Close()

	ok := resp.StatusCode < 400
	if opts.WantStatus != 0 {
		ok = resp.StatusCode == opts.WantStatus
	}
	detail := fmt.Sprintf("HTTP %d", resp.StatusCode)
	if !ok && opts.WantStatus != 0 {
		detail += fmt.Sprintf(" (attendu %d)", opts.WantStatus)
	}
	return Result{OK: ok, Detail: detail}
}

func checkTCP(addr string, opts Options) Result {
	_, port, err := net.SplitHostPort(addr)
	if err == nil {
		if p, perr := strconv.Atoi(port); perr != nil || p < 1 || p > 65535 {
			err = fmt.Errorf("port invalide %q", port)
		}
	}
	if err != nil {
		return Result{Detail: fmt.Sprintf("cible invalide (attendu http(s)://... ou host:port): %v", err)}
	}
	conn, err := net.DialTimeout("tcp", addr, opts.Timeout)
	if err != nil {
		return Result{Detail: describeErr(err)}
	}
	conn.Close()
	return Result{OK: true, Detail: "port ouvert"}
}

func describeErr(err error) string {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "délai dépassé"
	}
	return err.Error()
}
