package gooni

import (
	"flag"
	"log"
	"reflect"
	"strings"
	"testing"
)

const num_attempts = 3

var hostnames = flag.String("hostnames", "google.com", "comma-separated list of hostnames to test")
var resolvers = flag.String("resolvers", "8.8.8.8,8.8.4.4", "comma-separated list of resolvers to test")
var control_resolver = flag.String("control_resolver", "8.8.8.8", "resolver to use as control")

func TestDNSTamper(t *testing.T) {
	for _, hostname := range strings.Split(*hostnames, ",") {
		log.Printf("Testing %q against control %q", hostname, control_resolver)
		want, err := lookupIP(*control_resolver, hostname)
		if err != nil {
			t.Fatalf("Failed to resolve %q at control resolver %q", hostname, *control_resolver)
		}
		for _, resolver := range strings.Split(*resolvers, ",") {
			log.Printf("Testing %q against %q", hostname, resolver)
			got, err := lookupIP(resolver, hostname)
			if err != nil {
				t.Fatalf("Failed to resolve %q at resolver %q", hostname, resolver)
			}

			log.Printf("got %+v, want %+v", got, want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("want %q, got %q", want, got)
			}
		}
	}
}
