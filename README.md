## Vagrant - Nomad + Consul

Setup a 3 node nomad + consul server cluster & 2 nomad + docker clients on debian:12.   
Servers & clients are on a private subnet w/nat and secured w/mTLS.

Servers will be setup with ansible after the last client goes up.


## Network
* server01 - `192.168.60.10`
* server02 - `192.168.60.20`
* server03 - `192.168.60.30`
* client01 - `192.168.60.110`
* client02 - `192.168.60.120`

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
