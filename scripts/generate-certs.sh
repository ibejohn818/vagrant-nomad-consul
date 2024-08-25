#!/usr/bin/env sh
#
DIR=$(cd $(dirname "$0") && pwd)

# HELP="$(basename "$0") [OPTIONS] [SERVICE]
#
# Stage nomad or consul certs locally from AWS Secrets Manager.
# Also has action to remove staged certs.
#
# OPTIONS:
#      -h/--help  show this help text
#      --region the nomad region (not needed for consul)
#          --dc the datacenter
# ARGS:
#        SERVICE  nomad | consul (required)
#  "
# ARG_POS=0
# SERVICE=""
# DATACENTER=""
# REGION=""
#
# while [[ $# -gt 0 ]]; do
#   case $1 in
#     -h|--help)
#       echo "${HELP}"
#       exit 0;
#       ;;
#     --dc)
#       shift
#       DATACENTER="${1}"
#       shift
#       ;;
#     --region)
#       shift
#       REGION="${1}"
#       shift
#       ;;
#     -*|--*)
#       echo "Unknown flag $1"
#       exit 1
#       ;;
#     *)
#       case "${ARG_POS}" in
#         0)
#           SERVICE="${1}"
#           ARG_POS=1
#           shift
#           ;;
#         *)
#           echo "Unknown argument: ${1}"
#           echo ""
#           echo "${HELP}"
#           exit 1;
#       esac
#       ;;
#   esac
# done

set -ex

# if [[ -z "${DATACENTER}" ]]; then
#   echo "--dc cannot be empty"
#   exit 1
# fi
#
# if [[ "${SERVICE}" == "nomad" ]]; then
#   if [[ -z "${REGION}" ]]; then
#     echo "--region cannot be empty for nomad"
#     exit 1
#   fi
# fi

PWD=$(pwd)
mkdir -p "${PWD}/consul"
mkdir -p "${PWD}/nomad"
mkdir -p "${PWD}/service"

######################
# consul ca
#####################
#
openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 \
    -nodes -keyout consul-ca-key.pem -out consul-ca.pem -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab\ Consul\ CA" \
    -addext "subjectAltName=DNS:consul,DNS:localhost,IP:127.0.0.1" \
    -addext "basicConstraints=critical,CA:TRUE"

# consul server
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Consul" \
    -keyout  consul/server-key.pem \
    -out consul/server.csr \
    -addext "subjectAltName=DNS:consul.dc1.consul,DNS:server.dc1.consul,IP:127.0.0.1" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in consul/server.csr \
  -copy_extensions copy \
  -CA consul-ca.pem -CAkey consul-ca-key.pem -CAcreateserial \
  -out consul/server.pem

# consul client
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Consul" \
    -keyout  consul/client-key.pem \
    -out consul/client.csr \
    -addext "subjectAltName=DNS:consul.dc1.consul,DNS:client.dc1.consul,DNS:localhost,IP:127.0.0.1" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in consul/client.csr \
  -copy_extensions copy \
  -CA consul-ca.pem -CAkey consul-ca-key.pem -CAcreateserial \
  -out consul/client.pem

# create client PFX/P12
openssl pkcs12 -export \
  -out consul/client.p12 \
  -inkey consul/client-key.pem \
  -in consul/client.pem \
  -passout "pass:password" \
  -certfile consul-ca.pem

#################
# Nomad
################
openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 \
    -nodes -keyout nomad-ca-key.pem -out nomad-ca.pem -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab\ Nomad\ CA" \
    -addext "subjectAltName=DNS:nomad,DNS:localhost,IP:127.0.0.1" \
    -addext "basicConstraints=critical,CA:TRUE"

# nomad server
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Nomad" \
    -keyout  nomad/server-key.pem \
    -out nomad/server.csr \
    -addext "subjectAltName=DNS:nomad.dc1.consul,DNS:server.dc1.nomad,IP:127.0.0.1,IP:0.0.0.0" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in nomad/server.csr \
  -copy_extensions copy \
  -CA nomad-ca.pem -CAkey nomad-ca-key.pem -CAcreateserial \
  -out nomad/server.pem

# nomad client
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Nomad" \
    -keyout  nomad/client-key.pem \
    -out nomad/client.csr \
    -addext "subjectAltName=DNS:nomad.dc1.consul,DNS:client.dc1.nomad,DNS:localhost,IP:127.0.0.1,IP:0.0.0.0" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in nomad/client.csr \
  -copy_extensions copy \
  -CA nomad-ca.pem -CAkey nomad-ca-key.pem -CAcreateserial \
  -out nomad/client.pem

# create client PFX/P12
openssl pkcs12 -export \
  -out nomad/client.p12 \
  -inkey nomad/client-key.pem \
  -in nomad/client.pem \
  -passout "pass:password" \
  -certfile nomad-ca.pem

##################
# Service
#################

openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 \
    -nodes -keyout service-ca-key.pem -out service-ca.pem -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab\ Service\ CA" \
    -addext "subjectAltName=DNS:service,DNS:localhost,IP:127.0.0.1" \
    -addext "basicConstraints=critical,CA:TRUE"

# server
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Service Server" \
    -keyout  service/server-key.pem \
    -out service/server.csr \
    -addext "subjectAltName=DNS:*.service.local,DNS: service.local,DNS: *.service.dc1.consul, DNS: *.service.consul" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in service/server.csr \
  -copy_extensions copy \
  -CA service-ca.pem -CAkey service-ca-key.pem -CAcreateserial \
  -out service/server.pem

# client
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Service Client" \
    -keyout  service/client-key.pem \
    -out service/client.csr \
    -addext "subjectAltName=DNS:client.service.local" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 

openssl x509 -req -days 3650 \
  -in service/client.csr \
  -copy_extensions copy \
  -CA service-ca.pem -CAkey service-ca-key.pem -CAcreateserial \
  -out service/client.pem

# create client PFX/P12
openssl pkcs12 -export \
  -out service/client.p12 \
  -inkey service/client-key.pem \
  -in service/client.pem \
  -passout "pass:password" \
  -certfile service-ca.pem

