# PushProx [![CircleCI](https://circleci.com/gh/adobe/PushProx.svg?style=shield)](https://circleci.com/gh/adobe/PushProx)

PushProx is a client and proxy that allows transversing of NAT and other
similar network topologies by Prometheus, while still following the pull model.

While this is reasonably robust in practice, this is a work in progress.

## This Fork

This fork is a fork of github.com/robustperception/pushprox which changes the way in which clients scrape. Rather than
have the Prometheus pull determine where the client performs its final scrape from, the final scape URL is speficied
on the command line of the client. This ensures the client cant be used to scrape any host inside the network boundary where
the client is running. Other than that the changes are minimal, mostly Docker containers and more extensive help.


## Running

First build the proxy and client:

```
go get github.com/adobe/pushprox/{client,proxy}
cd ${GOPATH-$HOME/go}/src/github.com/adobe/pushprox/client
go build
cd ${GOPATH-$HOME/go}/src/github.com/adobe/pushprox/proxy
go build
```

Run the proxy somewhere both Prometheus and the clients can get to:

```
./proxy
```

On every target machine run the client, pointing it at the proxy:
```
./client --proxy-url=http://proxy:8080/ --pull-url=http://localhost:4502/metrics
```

In Prometheus, use the proxy as a `proxy_url`:

```
scrape_configs:
- job_name: node
  proxy_url: http://proxy:8080/
  static_configs:
    - targets: ['client:9100']  # Presuming the FQDN of the client is "client".
```

If the target must be scraped over SSL/TLS, add:
```
  params:
    _scheme: [https]
```
rather than the usual `scheme: https`. Only the default `scheme: http` works with the proxy,
so this workaround is required.

## Docker files

There are 2 Docker files. Dockerfile.client and Dockerfile.proxy for the client and proxy. The Proxy can
probably be run as is. The client should be used in annother Dockerfile containing the application copying 
the client binary into that Dockerfile and running the client as a background process with appropriate
command line params. Both --proxy-url and --pull-url must be specified, other parameters are optional.

To build the docker files
````
docker build -f Dockerfile.client .
docker build -f Dockerfile.proxy .
````

## Service Discovery

The `/clients` endpoint will return a list of all registered clients in the format
used by `file_sd_configs`. You could use wget in a cronjob to put it somewhere
file\_sd\_configs can read and then then relabel as needed.

## How It Works

The client registers with the proxy, and awaits instructions.

When Prometheus performs a scrape via the proxy, the proxy finds
the relevant client and tells it what to scrape. The client performs the scrape,
sends it back to the proxy which passes it back to Prometheus.

## Security

There is no authentication or authorisation included, a reverse proxy can be
put in front though to add these.

In the origial version, running the client allows those with access to the proxy or the client to access
all network services on the machine hosting the client. 

In this version, the pull url is hard coded on the command line and only allows the client to pull
from a fixed location.
