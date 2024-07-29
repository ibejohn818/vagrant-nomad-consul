## Vagrant - Nomad + Consul

Run a local HA nomad + consul server cluster with client nodes.

By default vagrant will setup 3 nomad & 3 consul server instances setup as HA cluster.   
In addition to the nomad + consul servers, a 3 node app client cluster will be created and a 
2 node data client cluster.   

Server OS is debian:12 and all services are configured VIA `ansible`.


## requirements
* vagrant
* hypervisor (IE: virtualbox | vmware)
* docker
* ansible 

| NOTE: tested with virtualbox on ubuntu & debian x86


## Network
Each instance will have a NAT nic to connect to the internet and a private network 
that nomad will utilize as its bridge network when creating docker tasks.

Private Net: `192.168.60.0\24`

### Servers
* nomad01-03: `192.168.60.11-13`
* consul01-03 - `192.168.60.20`

### Clients
* app01-03 - `192.168.60.30`
    * nomad meta
        - `class: app`
* data01-02 - `192.168.60.30`
    * nomad meta
        - `class: data`

### NOTE:
Nomad servers and clients are utilizing local consul clients for discovery.
Clients are also utilizing `dnsmasq` for `*.service.consul` dns resolution.

## Ansible

The `ansible` directory contains the ansible structure for all setup of the vagrant instances.   
The `Vagrantfile` will trigger ansible provisioning (playbook:`ansible/vagrant.yml`) on all instances in parallel after the last client has been booted.   

NOTE:   
After the initial booting and provisioning of the vagrant instances, the ansbile inventory can be copied to `ansible/inventory/vagrant/hosts` from `.vagrant/provisioning/ansible/` directory.


## Building the cluster

* start by generating the self-signed certificates

```shell
# via docker
make docker-generate-certs
# OR
# local openssl
make generate-certs
```
* startup the vagrant instances

```shell
vagrant up
```
## accessing nomad & consul

Nomad: (`nomad01-03` serve the dashboard)   
* nomad01 - http://192.168.60.11:4646
    

Consul: (`consul01-03` serve the dashboard) 
* consul01 - http://192.168.60.14:8501

For terminal access, first install the `nomad` & `consul` cli binaries.    

* nomad: https://developer.hashicorp.com/nomad/docs/install
* consul: https://developer.hashicorp.com/consul/docs/install

Source `nomad` env vars
```shell
# from the root of the repository
source scripts/nomad.env
# view the nomad servers and leader
nomad server members

```
Source `consul` env vars
```shell
TODO: Need to write
```


## Nomad job examples

 - TODO: Complete example docs

### TODO

* program an interface to build the cluster
* parameterize cluster options in `Vagrantfile` 
* add `podman` option for nomad clients
* 
