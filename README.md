## Vagrant - Nomad + Consul

Run a local HA nomad + consul server cluster with client nodes.

By default vagrant will setup 3 nomad & 3 consul server instances setup as HA cluster.   
In addition to the nomad + consul servers, 5 nomad client instances will be created.   

Server OS is `debian:12` and all services are configured VIA `ansible`.

## requirements
NOTE: tested on ubuntu & debian desktop 

* vagrant
* hypervisor (IE: virtualbox (tested) | vmware (vagrant plugin errors prevented testing) )
* docker
* ansible 

For terminal access, install the `nomad` & `consul` cli binaries.    

* nomad: https://developer.hashicorp.com/nomad/docs/install
* consul: https://developer.hashicorp.com/consul/docs/install

I also suggest running `dnsmasq` on the host machine to enable consul DNS resolution. (IE: {SERVICE-NAME}.service.dc1.consul)  
A configuration fragment for `dnsmasq` is located at `conf/dnsmasq.d/00-consul.conf`


## Network
Each instance will have a NAT nic to connect to the internet and a private network 
that nomad will utilize as its bridge network when creating docker tasks.

Private Net: `192.168.60.0\24`

### Servers
* nomad01-03: `192.168.60.11-13`
* consul01-03 - `192.168.60.14-16`

### Clients
* app01-03 - `192.168.60.21-23`
    * nomad meta
        - `class: app`
* data01-02 - `192.168.60.31-32`
    * nomad meta
        - `class: data`

### NOTE:
Nomad servers and clients are utilizing local consul clients for discovery.
Clients are also utilizing `dnsmasq` for `*.service.consul` dns resolution.

## Ansible

The `ansible` directory contains the ansible structure for all setup of the vagrant instances.   
The `Vagrantfile` will trigger ansible provisioning (playbook:`ansible/vagrant.yml`) on all instances in parallel after the last client has been started.   

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
After the instances have been provisioned, run the following script to add all certificates to the
consul KV.

```shell
scripts/consul-kv.sh
```

## accessing nomad & consul

Nomad: (`nomad01-03` serve the dashboard)   
* nomad01 - http://192.168.60.11:4646
    

Consul: (`consul01-03` serve the dashboard) 
* consul01 - http://192.168.60.14:8501

## nomad & consul cli

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
Nomad job examples reside in: `jobs/`.   
```shell
# source nomad env vars
source scripts/nomad.env

# execute a ingres job
cd jobs && nomad job run ingress.hcl
```
 - TODO: Complete example jobs doc

## Tooling

Tools for running nomad & consul at scale are located at `apps/`

#### `apps/gotooling` - golang tooling
    - `cmd/logs` - select job allocations logs to stream concurrently and/or write to file
    - `cmd/dev` - deploy and monitor multiple HCL's concurrently (TODO: rename binary)

#### `apps/rustooling`
    - TODO: docs on rust tooling
    
### Test hardware
Ubuntu 24.04 & Debian Bookworm
    - AMD 3950x 65gb ram
    - AMD 5950x 128gb ram
    - AMD 7950x 96gb ram
    - AMD 7960x 128gb ram
    - AMD 7975wx 128gb ram

### TODO

* program an interface to build the cluster
* parameterize cluster options in `Vagrantfile` 
* add `podman` option for nomad clients
* 
