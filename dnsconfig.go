package gooni

type dnsConfig struct {
	resolver string	  // server to use
	timeout  int      // seconds before giving up on packet
	ndots	 int	  // number of dots to trigger absolute lookup
	attempts int      // lost packets before giving up on server
}

func dnsCreateConfig(resolver string) (*dnsConfig, error) {
	conf := new(dnsConfig)
	conf.resolver = resolver
	conf.timeout = 5
	conf.ndots = 2
	conf.attempts = 2

	return conf, nil
}
