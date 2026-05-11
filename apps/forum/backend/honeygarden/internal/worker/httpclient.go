package worker

import (
	"slices"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"syscall"
	"time"
)

var ErrBlockedAddress = errors.New("blocked internal address")

// ValidatePublicURL проверяет scheme и DNS-резолв, отбрасывая приватные сети.
// Используется при регистрации источника, чтобы не сохранять заведомо SSRF-URL.
func ValidatePublicURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("missing host")
	}
	addrs, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip", host)
	if err != nil {
		return fmt.Errorf("dns lookup: %w", err)
	}
	if slices.ContainsFunc(addrs, isBlockedAddr) {
			return ErrBlockedAddress
		}
	return nil
}

func isBlockedAddr(a netip.Addr) bool {
	if !a.IsValid() {
		return true
	}
	return a.IsLoopback() ||
		a.IsPrivate() ||
		a.IsLinkLocalUnicast() ||
		a.IsLinkLocalMulticast() ||
		a.IsMulticast() ||
		a.IsUnspecified() ||
		a.IsInterfaceLocalMulticast()
}

// NewSafeHTTPClient создаёт http.Client с транспортом, отбраковывающим
// соединения с приватными/loopback/link-local IP. Защита срабатывает
// и после DNS-ребиндинга, и на редиректах
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			addr, err := netip.ParseAddr(host)
			if err != nil {
				return fmt.Errorf("parse addr: %w", err)
			}
			if isBlockedAddr(addr) {
				return ErrBlockedAddress
			}
			return nil
		},
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to unsupported scheme %q", req.URL.Scheme)
			}
			return nil
		},
	}
}
