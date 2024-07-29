#!/usr/bin/env bash

DIR=$(cd $(dirname "${0}") && pwd)
TLS_DIR="${DIR}/../tls"

function consul_cmd() {
  CONSUL_HTTP_ADDR=https://192.168.60.14:8501 \
  CONSUL_CACERT="${TLS_DIR}/consul-ca.pem" \
  CONSUL_CLIENT_CERT="${TLS_DIR}/consul/client.pem" \
  CONSUL_CLIENT_KEY="${TLS_DIR}/consul/client-key.pem" \
  CONSUL_HTTP_SSL_VERIFY="false" \
  consul ${@}
}


function consul_chk() {
  local key="${1}"
  consul_cmd kv get "${key}" &>/dev/null
  echo "${?}"
}

function consul_file() {
  local key="${1}"
  local path="${2}"

  CHK=$(consul_chk "${key}")

  if [[ "${CHK}" == "1" ]]; then
    consul_cmd kv put "${key}" "@${path}"
  fi

}


consul_file "tls/consul/ca.pem" "${TLS_DIR}/consul-ca.pem"
consul_file "tls/consul/ca.key" "${TLS_DIR}/consul-ca-key.pem"
consul_file "tls/consul/server.pem" "${TLS_DIR}/consul/server.pem"
consul_file "tls/consul/server.key" "${TLS_DIR}/consul/server-key.pem"
consul_file "tls/consul/client.pem" "${TLS_DIR}/consul/client.pem"
consul_file "tls/consul/client.key" "${TLS_DIR}/consul/client-key.pem"

consul_file "tls/nomad/ca.pem" "${TLS_DIR}/nomad-ca.pem"
consul_file "tls/nomad/ca.key" "${TLS_DIR}/nomad-ca-key.pem"
consul_file "tls/nomad/server.pem" "${TLS_DIR}/nomad/server.pem"
consul_file "tls/nomad/server.key" "${TLS_DIR}/nomad/server-key.pem"
consul_file "tls/nomad/client.pem" "${TLS_DIR}/nomad/client.pem"
consul_file "tls/nomad/client.key" "${TLS_DIR}/nomad/client-key.pem"

consul_file "tls/service/ca.pem" "${TLS_DIR}/service-ca.pem"
consul_file "tls/service/ca.key" "${TLS_DIR}/service-ca-key.pem"
consul_file "tls/service/server.pem" "${TLS_DIR}/service/server.pem"
consul_file "tls/service/server.key" "${TLS_DIR}/service/server-key.pem"
consul_file "tls/service/client.pem" "${TLS_DIR}/service/client.pem"
consul_file "tls/service/client.key" "${TLS_DIR}/service/client-key.pem"
