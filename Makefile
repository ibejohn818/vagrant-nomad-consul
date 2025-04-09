DOCKER_REPO ?= ibejohn818/nomad-consul-vagrant

build-openssl-image:
	$(MAKE) -C docker build-openssl

docker-generate-certs: build-openssl-image
	docker run --rm -it \
		--user $(shell id -u):$(shell id -g) \
		-v /etc/passwd:/etc/passwd:ro \
		-v $(shell pwd)/tls:/tls \
		-v $(shell pwd)/scripts/generate-certs.sh:/generate-certs.sh \
		--workdir /tls \
		$(DOCKER_REPO):openssl \
		/generate-certs.sh

generate-certs:
	cd tls && ../scripts/generate-certs.sh


