package main

import (
	"net"
	"net/http"
	"regexp"
)

// MatchesRule checks all rules in the conf and tries to match on them
func (c Conf) MatchesRule(r *http.Request) (Rule, string, bool) {
	ip := parseIP(r)
	ua := parseUA(r)

	for _, rule := range c.Rules {
		matchedUA, _ := regexp.Match(rule.UARegex, []byte(ua))
		if matchedUA && rule.Mode == "ua" {
			return rule, ua, true
		}

		if matchedUA && rule.Mode == "ip+ua" {
			return rule, ip + ua, true
		}

		if rule.Mode == "ip" {
			return rule, ip, true
		}
	}
	return Rule{}, "", false
}

// To prevent IP spoofing, be sure to delete any pre-existing
// X-Forwarded-For header coming from the client or
// an untrusted proxy.

func parseIP(r *http.Request) string {
	if xdi := r.Header.Get("X-Device-IP"); xdi != "" {
		return ip(xdi)
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return ip(xff)
	}
	return ip(r.RemoteAddr)
}

func ip(s string) string {
	ip, _, err := net.SplitHostPort(s)
	if err != nil {
		return s
	}
	return ip
}

func parseUA(r *http.Request) string {
	if xdua := r.Header.Get("X-Device-User-Agent:"); xdua != "" {
		return xdua
	}

	if ua := r.Header.Get("User-Agent"); ua != "" {
		return ua
	}

	return "[err-no-ua]"
}
