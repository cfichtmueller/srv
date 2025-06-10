// Copyright 2025 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package srv

import (
	"net"
	"net/http"
	"strings"
)

type IPResolver struct {
	RemoteIPHeaders      []string
	TrustRemoteIdHeaders bool
}

func NewIPResolver(remoteIPHeaders []string, trustRemoteIdHeaders bool) *IPResolver {
	return &IPResolver{
		RemoteIPHeaders:      remoteIPHeaders,
		TrustRemoteIdHeaders: trustRemoteIdHeaders,
	}
}

func (r *IPResolver) Resolve(req *http.Request) []string {
	remoteIP := getRemoteIP(req)
	if !r.TrustRemoteIdHeaders || len(r.RemoteIPHeaders) == 0 {
		return []string{remoteIP}
	}
	ips := make([]string, 0, 2)
	for _, headerName := range r.RemoteIPHeaders {
		headerValue := req.Header.Get(headerName)
		if headerValue == "" {
			continue
		}
		switch headerName {
		case "X-Forwarded-For":
			rawIPs := strings.Split(headerValue, ",")
			for _, rawIP := range rawIPs {
				ip := strings.TrimSpace(rawIP)
				if net.ParseIP(ip) != nil {
					ips = append(ips, ip)
				}
			}
		}
	}
	if len(ips) == 0 || remoteIP != ips[len(ips)-1] {
		ips = append(ips, remoteIP)
	}
	return ips
}

func getRemoteIP(req *http.Request) string {
	rawIP, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err != nil {
		return ""
	}
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return ""
	}
	return ip.String()
}
