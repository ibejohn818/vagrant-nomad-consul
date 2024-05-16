
REPO ?= ibejohn818/nomad-consul-vagrant

build-openssl:
	docker buildx build \
		-t $(REPO):openssl \
		openssl

push-openssl:
	docker push $(REPO):openssl
