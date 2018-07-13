

TAGBASE = hub.docker.com/something

VERSION = 1.2

default: build


build: client/client proxy/proxy discovery/discovery

client/client: client/*.go 
	echo "Building client"
	cd client && go get && go build

proxy/proxy: proxy/*.go
	echo "Building proxy"
	cd proxy && go get && go build

discovery/discovery: discovery/*.go
	echo "Building discovery"
	cd discovery && go get && go build


push: docker client_push proxy_push discovery_push

docker: proxy_image client_image discovery_image

client_image:
	docker build -t $(TAGBASE)/client:$(VERSION) -f Dockerfile.client .

proxy_image:
	docker build -t $(TAGBASE)/proxy:$(VERSION) -f Dockerfile.proxy .

discovery_image:
	docker build -t $(TAGBASE)/discovery:$(VERSION) -f Dockerfile.discovery .

client_push:
	docker build $(TAGBASE)/client:$(VERSION)

proxy_push:
	docker build $(TAGBASE)/proxy:$(VERSION)

discovery_push:
	docker push $(TAGBASE)/discovery:$(VERSION) 

