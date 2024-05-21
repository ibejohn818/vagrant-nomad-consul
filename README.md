## Vagrant - Nomad + Consul

Setup a 3 node nomad + consul server cluster & 2 nomad + docker clients on debian:12.   
Servers & clients are on a private subnet w/nat and secured w/mTLS.

Note:
Nomad + Consul can be setup on the same host's to lower VM count. (Reference the vagrant file)

Servers will be setup with ansible after the last client goes up.


## requirements
* vagrant
* hypervisor (virtualbox / vmware / kvm .. the latter 2 require vagrant plugins)
* docker
* make

## Building the cluster

* start by generating the self-signed certificates

```shell
make docker-generate-certs

```
