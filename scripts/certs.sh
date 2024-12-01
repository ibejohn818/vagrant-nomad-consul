#!/usr/bin/env bash

HERE=$(cd $(dirname "$0") && pwd)

# ensure we're being run from the root of the repo
if [[ ! $(stat "${HERE}/.git") ]]; then
  echo ".git dir not found! "
  echo "script must be executed from the root of the repo"
  exit 1;
fi




HELP="$(basename "$0") [OPTIONS] [SERVICE]

Stage nomad or consul certs locally from AWS Secrets Manager.
Also has action to remove staged certs.

OPTIONS:
     -h/--help  show this help text
     --region the nomad region (not needed for consul)
         --dc the datacenter
ARGS:
       SERVICE  nomad | consul (required)
 "

EXP_DAYS=3650
ARG_POS=0
CERT_TYPE=""
DATACENTER=""
REGION=""
SAVE_TO="${HERE}/tls"

while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--help)
      echo "${HELP}"
      exit 0;
      ;;
    --dc)
      shift
      DATACENTER="${1}"
      shift
      ;;
    --region)
      shift
      REGION="${1}"
      shift
      ;;
    -*|--*)
      echo "Unknown flag $1"
      exit 1
      ;;
    *)
      case "${ARG_POS}" in
        0)
          CERT_TYPE="${1}"
          ARG_POS=1
          shift
          ;;
        *)
          echo "Unknown argument: ${1}"
          echo ""
          echo "${HELP}"
          exit 1;
      esac
      ;;
  esac
done

# validate parameters

if [[ -z "${CERT_TYPE}" ]]; then 
  echo "argument for CERT_TYPE cannot be empty"
  exit 1
fi

save_path=$"${SAVE_TO}/${CERT_TYPE}"
ca_key="${CERT_TYPE}-ca-key.pem"
ca="${CERT_TYPE}-ca.pem"

# always ensure path
mkdir -p "${save_path}"

function gen_ca() {

  openssl req -x509 -newkey rsa:4096 -sha256 -days ${EXP_DAYS} \
    -nodes -keyout "${ca_key}" -out "${ca}" -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab\ ${CERT_TYPE}\ CA" \
    -addext "subjectAltName=DNS:${CERT_TYPE},DNS:localhost,IP:127.0.0.1" \
    -addext "basicConstraints=critical,CA:TRUE"

}



function gen_cert() {

# csr
openssl x509 -req -days 3650 \
  -in consul/server.csr \
  -copy_extensions copy \
  -CA consul-ca.pem -CAkey consul-ca-key.pem -CAcreateserial \
  -out consul/server.pem

# cert
openssl req -new -newkey rsa:4096 -nodes  \
    -subj "/C=US/ST=CA/O=Lab/OU=Engineering/CN=Lab Consul" \
    -keyout  consul/client-key.pem \
    -out consul/client.csr \
    -addext "subjectAltName=DNS:consul.dc1.consul,DNS:client.dc1.consul,DNS:localhost,IP:127.0.0.1" \
    -addext "extendedKeyUsage = serverAuth, clientAuth" \
    -addext "basicConstraints=critical,CA:FALSE" 
}
