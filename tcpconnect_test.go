package gooni

import (
	"flag"
	"net"
	"net/url"
	"strings"
	"testing"
)

var endpoints = flag.String("endpoints", "http://www.google.com,google.com:80,8.8.8.8:53", "comma-separated list of endpoints to attempt to connect to")

func TestTCPConnect(t *testing.T) {
	for _, endpoint := range strings.Split(*endpoints, ",") {
		t.Logf("Testing %q", endpoint)
		_, err := net.Dial("tcp", endpoint)
		if err == nil {
			continue
		}
		u, err := url.Parse(endpoint)
		if err != nil {
			t.Fatalf("Failed to parse %q as URL", endpoint)
		}
		t.Logf("Trying as URL %+v", *u)
		endpoint = u.Host
		if !strings.Contains(endpoint, ":") {
			// Set port from scheme
			switch u.Scheme {
				case "http", "https":
					endpoint += ":80"
				default:
					t.Fatalf("Unknown scheme %q", u.Scheme)
			}
		}
		_, err = net.Dial("tcp", endpoint)
		if err != nil {
			t.Errorf("Failed to connect to %q", endpoint)
		}
	}
}
