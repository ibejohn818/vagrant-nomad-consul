
REPO ?= ibejohn818/nomad-consul-vagrant

PLATFORM ?= linux/amd64
REGISTRY ?= registry.service.dc1.consul
TAG ?= latest

.PHONY: postgres

build-openssl:
	docker buildx build \
		-t $(REPO):openssl \
		openssl

push-openssl:
	docker push $(REPO):openssl

postgres:
	docker buildx build \
		--platform $(PLATFORM) \
		-t postgres-ha:$(TAG) \
		postgres 


tag_%:
	docker tag $*:$(TAG) $(REGISTRY)/$*:$(TAG)

push_%:
	docker push --disable-content-trust $(REGISTRY)/$*:$(TAG)
