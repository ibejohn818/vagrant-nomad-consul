job "ingress" {

    datacenters = ["global"]
    region = "dc1"
    type = "service"
    group "server" {

        count = 5
        constraint {
            attribute = "${node.class}"
      operator="set_contains_any"
            value = "app,data"
        }

        network {
            port "http" {
                static = 80
            }
            port "https" {
                static = 443
            }
            port "api" {
                static = 8080
            }
        }

        service  {
            name  = "traefik-ssl"
            port = "https"
            check {
                type = "tcp"
                port = "https"
                interval = "15s"
                timeout = "5s"
            }
        }
        service  {
            name  = "traefik"
            port = "http"
            check {
                type = "tcp"
                port = "http"
                interval = "15s"
                timeout = "5s"
            }
        }

        task "traefik" {

            template {
                data = "{{ key \"tls/consul/ca.pem\" }}"
                destination = "secrets/consul/ca.pem"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/consul/client.pem\" }}"
                destination = "secrets/consul/client.pem"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/consul/client.key\" }}"
                destination = "secrets/consul/client.key"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/nomad/ca.pem\" }}"
                destination = "secrets/nomad/ca.pem"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/nomad/client.key\" }}"
                destination = "secrets/nomad/client.key"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/nomad/client.pem\" }}"
                destination = "secrets/nomad/client.pem"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/service/ca.pem\" }}"
                destination = "secrets/service/ca.pem"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/service/server.pem\" }}"
                destination = "secrets/service/server.pem"
                change_mode = "noop"
            }
            template {
                data = "{{ key \"tls/service/server.key\" }}"
                destination = "secrets/service/server.key"
                change_mode = "noop"
            }
            template {
                data = "{{ key \"tls/service/client.pem\" }}"
                destination = "secrets/service/client.pem"
                change_mode = "noop"
             }

            template {
                data = <<EOF
tls:
  stores:
    default:
      defaultCertificate:
        certFile: /secrets/service/server.pem
        keyFile: /secrets/service/server.key
  options:
    default:
      minVersion: VersionTLS12
    mtls:
      clientAuth:
        clientAuthType: RequireAndVerifyClientCert
        caFiles:
          - /secrets/service/ca.pem
  certificates:
      certFile: /secrets/service/server.pem
      keyFile: /secrets/service/server.key

EOF
                destination = "/local/conf/default.yaml"
                change_mode = "noop"
            }

            template {
                data = <<EOF
http:
  routers:
    dashboard:
      rule: Host(`traefik.service.dc1.consul`)
      service: api@internal
      tls: true

EOF
                destination = "/local/conf/dashboard.yaml"
                change_mode = "noop"
            }
            template {
                data = <<EOF
tcp:
  routers:
    nomad:
      rule: HostSNI(`nomad.service.dc1.consul`)
      tls:
          passthrough: true
      service: nomad-svc
  services:
    nomad-svc:
      loadbalancer:
        servers:{{range service "http.nomad"}}
          - address: "{{.Address}}:{{.Port}}"{{end}}

EOF
                destination = "/local/conf/nomad.yaml"
                change_mode = "noop"
            }



            template {
                data = <<EOF
tcp:
    routers:
        consul:
            rule: HostSNI(`consul.service.dc1.consul`)
            tls:
                passthrough: true
            service: consul-svc
    services:
        consul-svc:
            loadbalancer:
                servers:
                    - address: 127.0.0.1:8501
EOF
                destination = "/local/conf/consul.yaml"
                change_mode = "noop"
            }

            driver = "docker"

            config {

                image = "traefik:v2.11"

                args = [
                    "--entryPoints.http.address=:80",
                    "--entryPoints.https.address=:443",
                    "--accessLog.filePath=/dev/stdout",
                    "--api.dashboard=true",
                    "--api.insecure=true",
                    "--log.level=INFO",
                    "--providers.docker",
                    "--providers.file.directory=/local/conf",
                    "--providers.file.watch=true",
                    "--providers.consulCatalog.prefix=traefik",
                    "--providers.consulCatalog.refreshInterval=35s",
                    "--providers.consulCatalog.exposedByDefault=false",
                    "--providers.consulCatalog.endpoint.address=127.0.0.1:8501",
                    "--providers.consulCatalog.endpoint.scheme=https",
                    "--providers.consulCatalog.endpoint.tls.ca=/secrets/consul/ca.pem",
                    "--providers.consulCatalog.endpoint.tls.cert=/secrets/consul/client.pem",
                    "--providers.consulCatalog.endpoint.tls.key=/secrets/consul/client.key",
                    "--providers.consulCatalog.stale=false"
                ]

                network_mode = "host"
                ports = [
                    "http",
                    "https",
                    "api"
                ]
                privileged = true
                volumes = [
                    "/var/run/docker.sock:/var/run/docker.sock"
                ]
            }
        }
    }
}
