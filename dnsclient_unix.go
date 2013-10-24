// DNS client: see RFC 1035.

package gooni

import (
	"math/rand"
	"net"
	"time"
)

// Send a request on the connection and hope for a reply.
// Up to cfg.attempts attempts.
func exchange(cfg *dnsConfig, c net.Conn, name string, qtype uint16) (*dnsMsg, error) {
	if len(name) >= 256 {
		return nil, &DNSError{Err: "name too long", Name: name}
	}
	out := new(dnsMsg)
	out.id = uint16(rand.Int()) ^ uint16(time.Now().UnixNano())
	out.question = []dnsQuestion{
		{name, qtype, dnsClassINET},
	}
	out.recursion_desired = true
	msg, ok := out.Pack()
	if !ok {
		return nil, &DNSError{Err: "internal error - cannot pack message", Name: name}
	}

	for attempt := 0; attempt < cfg.attempts; attempt++ {
		n, err := c.Write(msg)
		if err != nil {
			return nil, err
		}

		if cfg.timeout == 0 {
			c.SetReadDeadline(time.Time{})
		} else {
			c.SetReadDeadline(time.Now().Add(time.Duration(cfg.timeout) * time.Second))
		}

		buf := make([]byte, 2000) // More than enough.
		n, err = c.Read(buf)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			}
			return nil, err
		}
		buf = buf[0:n]
		in := new(dnsMsg)
		if !in.Unpack(buf) || in.id != out.id {
			continue
		}
		return in, nil
	}
	var server string
	if a := c.RemoteAddr(); a != nil {
		server = a.String()
	}
	return nil, &DNSError{Err: "no answer from server", Name: name, Server: server, IsTimeout: true}
}

// Do a lookup for a single name, which must be rooted
// (otherwise answer will not find the answers).
func tryOneName(cfg *dnsConfig, name string, qtype uint16) (cname string, addrs []dnsRR, err error) {
	// Calling Dial here is scary -- we have to be sure
	// not to dial a name that will require a DNS lookup,
	// or Dial will call back here to translate it.
	// The DNS config parser has already checked that
	// all the cfg.servers[i] are IP addresses, which
	// Dial will use without a DNS lookup.
	server := cfg.resolver + ":53"
	c, cerr := net.Dial("udp", server)
	if cerr != nil {
		err = cerr
		return
	}
	msg, merr := exchange(cfg, c, name, qtype)
	c.Close()
	if merr != nil {
		err = merr
		return
	}
	cname, addrs, err = answer(name, server, msg, qtype)
	if err == nil || err.(*DNSError).Err == noSuchHost {
		return
	}
	return
}

func convertRR_A(records []dnsRR) []net.IP {
	addrs := make([]net.IP, len(records))
	for i, rr := range records {
		a := rr.(*dnsRR_A).A
		addrs[i] = net.IPv4(byte(a>>24), byte(a>>16), byte(a>>8), byte(a))
	}
	return addrs
}

func convertRR_AAAA(records []dnsRR) []net.IP {
	addrs := make([]net.IP, len(records))
	for i, rr := range records {
		a := make(net.IP, net.IPv6len)
		copy(a, rr.(*dnsRR_AAAA).AAAA[:])
		addrs[i] = a
	}
	return addrs
}

func count (s string, b byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			n++
		}
	}
	return n
}

func lookup(cfg *dnsConfig, name string, qtype uint16) (cname string, addrs []dnsRR, err error) {
	if !isDomainName(name) {
		return name, nil, &DNSError{Err: "invalid domain name", Name: name}
	}
	// If name is rooted (trailing dot) or has enough dots,
	// try it by itself first.
	rooted := len(name) > 0 && name[len(name)-1] == '.'
	if rooted || count(name, '.') >= cfg.ndots {
		rname := name
		if !rooted {
			rname += "."
		}
		// Can try as ordinary name.
		cname, addrs, err = tryOneName(cfg, rname, qtype)
		if err == nil {
			return
		}
	}
	if rooted {
		return
	}

	// Last ditch effort: try unsuffixed.
	rname := name
	if !rooted {
		rname += "."
	}
	cname, addrs, err = tryOneName(cfg, rname, qtype)
	if err == nil {
		return
	}
	return
}

func lookupIP(resolver, name string) (addrs []net.IP, err error) {
	cfg, dnserr := dnsCreateConfig(resolver)
	if dnserr != nil || cfg == nil {
		err = dnserr
		return
	}
	var records []dnsRR
	var cname string
	cname, records, err = lookup(cfg, name, dnsTypeA)
	if err != nil {
		return
	}
	addrs = convertRR_A(records)
	if cname != "" {
		name = cname
	}
	_, records, err = lookup(cfg, name, dnsTypeAAAA)
	if err != nil && len(addrs) > 0 {
		// Ignore error because A lookup succeeded.
		err = nil
	}
	if err != nil {
		return
	}
	addrs = append(addrs, convertRR_AAAA(records)...)
	return
}

