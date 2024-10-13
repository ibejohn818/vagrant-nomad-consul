variable "version" {
    type = string
    default = "11.1.4-ubuntu"
}

variable "host_placement" {
  type = string
  default = "data02"
}

job "grafana" {
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
          port "http" {to = 3000 }
        }

        service {

            name = "grafana"
            tags = [
                "grafana",
                "traefik.enable=true",
                "traefik.http.routers.grafana.rule=Host(`grafana.service.dc1.consul`) || Host(`grafana.vagrant.local`)",
                "traefik.http.routers.grafana.entrypoints=http",
                "traefik.http.routers.grafana.middlewares=grafana-https",
                "traefik.http.middlewares.grafana-https.redirectscheme.scheme=https",

                "traefik.http.routers.grafanassl.rule=Host(`grafana.service.dc1.consul`) || Host(`grafana.vagrant.local`)",
                "traefik.http.routers.grafanassl.entrypoints=https",
                "traefik.http.routers.grafanassl.tls=true",
            ]

            port = "http"

        }

        task "grafana" {

            template {
                data = <<EOH
[database]
type = sqlite3
path = /data/db/grafana.sqlite

[paths]
data = /data
logs = /dev/stderr

[security]
admin_user     = admin

[metrics]
enabled             = false
disable_total_stats = true
EOH
                destination = "/local/config/grafana.ini"
                change_mode = "noop"
            }

            driver = "docker"

            env {
                GF_INSTALL_PLUGINS = "grafana-piechart-panel"
                GF_PATHS_DATA = "/data"
            }

            user = "root"
            config {
                privileged=true
                image = "grafana/grafana:${var.version}"
                ports = ["http"]
                args = [ "--config", "/local/config/grafana.ini"]

                dns_servers = [
                  "172.17.0.1"
                ]

               mount {
                  type = "bind"
                  target = "/data"
                  source = "/vagrant/data/grafana"
                  readonly = false
                }
            }

            resources {
                memory=1024
                disk=1024
                cpu=1000
            }
        
        }

    
    }
}
