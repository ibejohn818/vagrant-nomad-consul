variable "host_placement" {
  type = string
  default = "data01"
}

job "prometheus" {
    datacenters = ["global"]
    region = "dc1"
    type = "service"

    constraint {
      attribute = "${node.class}"
      operator = "="
      value = "data"
    }

    constraint {
      attribute = "${attr.unique.hostname}"
      operator = "="
      value = "${var.host_placement}"
    }

    group "server" {
        network {
          port "http" {to = 9090 }
        }

        service {

            name = "prometheus"

            port = "http"

            # check {
            #     type     = "tcp"
            #     port     = "http"
            #     interval = "30s"
            #     timeout  = "5s"
            # }

            tags = [
                "prometheus",
                "traefik.enable=true",
                "traefik.http.routers.promssl.rule=Host(`prometheus.service.dc1.consul`)",
                "traefik.http.routers.promssl.entrypoints=https",
                "traefik.http.routers.promssl.tls=true",
                //"traefik.http.routers.promssl.tls.options=mtls@file"
            ]
            
        }

        task "prometheus" {
            template {
                data = file("includes/prometheus.yml")
                destination = "/local/config/prom.yml"
            }
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

            driver = "docker"

            env {
            }

            config {

                args = [
                    "--config.file=/local/config/prom.yml",
                    "--web.listen-address=:${NOMAD_PORT_http}"
                ]
                network_mode = "host"
                privileged=true
                image = "prom/prometheus:v2.47.2"
                ports = ["http"]
                mounts = [
                    {
                        type     = "volume"
                        target   = "/prometheus"
                        source   = "prometheus-data"
                        readonly = false
                    }
                ]
            }
            resources {
                memory=1024
                cpu=1800
            }
        
        }

    
    }
}
