gooni
=====

go port of ooni.
        $ go test -v

dnstamper
---------
        $ go test -run TestDNSTamper --hostnames=<comma-separated list of hostnames> --resolvers=<comma-separated list of resolvers> --control_resolver=<control resolver>

tcpconnect
----------
        $ go test -run TestTCPConnect --endpoints=<comma-separated list of endpoints>
