// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"net/http"
	"testing"
)

func TestIPResolver_Resolve_NoHeaders(t *testing.T) {
	resolver := NewIPResolver(nil, false)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	ips := resolver.Resolve(req)

	if len(ips) != 1 {
		t.Errorf("Expected 1 IP, got %d", len(ips))
	}
	if ips[0] != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", ips[0])
	}
}

func TestIPResolver_Resolve_WithHeaders_NotTrusted(t *testing.T) {
	resolver := NewIPResolver([]string{"X-Forwarded-For"}, false)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	ips := resolver.Resolve(req)

	if len(ips) != 1 {
		t.Errorf("Expected 1 IP, got %d", len(ips))
	}
	if ips[0] != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", ips[0])
	}
}

func TestIPResolver_Resolve_WithHeaders_Trusted(t *testing.T) {
	resolver := NewIPResolver([]string{"X-Forwarded-For"}, true)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2, 192.168.1.1")

	ips := resolver.Resolve(req)

	if len(ips) != 3 {
		t.Errorf("Expected 3 IPs, got %d", len(ips))
		return
	}
	expectedIPs := []string{"10.0.0.1", "10.0.0.2", "192.168.1.1"}
	for i, expected := range expectedIPs {
		if ips[i] != expected {
			t.Errorf("Expected IP %s at position %d, got %s", expected, i, ips[i])
		}
	}
}

func TestIPResolver_Resolve_WithHeaders_Trusted_RemoteMissing(t *testing.T) {
	resolver := NewIPResolver([]string{"X-Forwarded-For"}, true)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	ips := resolver.Resolve(req)

	if len(ips) != 3 {
		t.Errorf("Expected 3 IPs, got %d", len(ips))
		return
	}
	expectedIPs := []string{"10.0.0.1", "10.0.0.2", "192.168.1.1"}
	for i, expected := range expectedIPs {
		if ips[i] != expected {
			t.Errorf("Expected IP %s at position %d, got %s", expected, i, ips[i])
		}
	}
}

func TestIPResolver_Resolve_InvalidRemoteAddr(t *testing.T) {
	resolver := NewIPResolver(nil, false)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "invalid"

	ips := resolver.Resolve(req)

	if len(ips) != 1 {
		t.Errorf("Expected 1 IP, got %d", len(ips))
	}
	if ips[0] != "" {
		t.Errorf("Expected empty IP, got %s", ips[0])
	}
}

func TestIPResolver_Resolve_InvalidHeaderIP(t *testing.T) {
	resolver := NewIPResolver([]string{"X-Forwarded-For"}, true)
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("X-Forwarded-For", "invalid, 10.0.0.2")

	ips := resolver.Resolve(req)

	if len(ips) != 2 {
		t.Errorf("Expected 2 IPs, got %d", len(ips))
		return
	}
	expectedIPs := []string{"10.0.0.2", "192.168.1.1"}
	for i, expected := range expectedIPs {
		if ips[i] != expected {
			t.Errorf("Expected IP %s at position %d, got %s", expected, i, ips[i])
		}
	}
}
