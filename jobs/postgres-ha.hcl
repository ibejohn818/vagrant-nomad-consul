variable "node_class" {
    type = string
    default = "app"
}

variable "db_port" {
  type = number
  default = 5432
}

variable "db_rest_port" {
    type = number
    default = 8008
}

variable "namespace" {
  type = string
  default = null
}

variable "count" {
  type    = number
  default = 2
}

variable "postgres_cpu" {
  type = string
  description = "Nomad CPU Allocation"
  default = "500"
}

variable "postgres_memory" {
  type = string
  description = "Nomad Memory Allocation"
  default = "256"
}

variable "postgres_tag" {
  type = string
  description = "image tag"
  default = "latest"
}

variable "postgres_exporter_cpu" {
  type = string
  description = "Nomad CPU Allocation"
  default = "2000"
}

variable "postgres_exporter_memory" {
  type = string
  description = "Nomad Memory Allocation"
  default = "256"
}

variable "max_connections" {
    type = number
    default = 200
}

job "postgres-ha" {
    datacenters = ["global"]
    type = "service"


    meta {
        db_rest_port       = var.db_rest_port
        db_port            = var.db_port
        db_max_connections = var.max_connections
    }

    update {
        max_parallel      = 1
        min_healthy_time  = "1m"
        healthy_deadline  = "120m"
        progress_deadline = "130m"
    }

    group "postgres" {

        network {
            port "postgres" {
                static = var.db_port
            }
            port "postgres-exporter" {
                to = 9187
            }
        }

        constraint {
            operator  = "distinct_hosts"
            value     = "true"
        }

        constraint {
            attribute = "${node.class}"
            operator = "regexp"
            value = "${var.node_class}"
        }

        service {
            name = "postgres-ha-http-api"
            port = "${var.db_rest_port}"
            check {
                name     = "postgres-ha-http-api"
                type     = "http"
                interval = "5s"
                timeout  = "2s"
                path     = "/readiness"
            }
        }

        service {
            name = "postgres-exporter"
            port = "postgres-exporter"
            tags = [ "metrics" ]
            check {
                name = "postgres-exporter-tcp"
                type = "tcp"
                interval = "5s"
                timeout = "2s"
            }
        }

        restart {
            attempts = 3
            delay    = "15s"
            interval = "1m"
            mode     = "delay"
        }

        reschedule {
            delay = "15s"
            delay_function = "constant"
            unlimited = true
        }

        count = var.count

        task "postgres" {
            driver = "docker"

            template {
                data = "{{ key \"tls/consul/client.pem\" }}"
                destination = "secrets/consul/client.crt"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/consul/client.key\" }}"
                destination = "secrets/consul/client.key"
                change_mode = "noop"
            }

            template {
                data = "{{ key \"tls/consul/ca.pem\" }}"
                destination = "secrets/consul/ca.crt"
                change_mode = "noop"
            }

            template {
                data = file("includes/patroni.yaml")
                destination = "/local/config/postgresql.yaml"
                change_mode = "noop"
            }

            config {
                image = "registry.service.dc1.consul/postgres-ha:${var.postgres_tag}"
                network_mode = "host"
                args = [ "patroni", "/local/config/postgresql.yaml" ]
                ulimit {
                    nproc = "1024"
                    nofile = "70000:70000"
                }
            }

            resources {
                cpu = "${var.postgres_cpu}"
                memory = "${var.postgres_memory}"
            }
        }


        task "postgres-exporter" {
            driver = "docker"

            env {
                PG_EXPORTER_AUTO_DISCOVER_DATABASES = "true"
                PG_EXPORTER_EXTEND_QUERY_PATH = "/local/config/postgres-exporter.yaml"
            }

            template {
              data = <<EOF
DATA_SOURCE_NAME = 'postgresql://postgres:somepassword@{{ env "attr.unique.network.ip-address" }}:{{ env "NOMAD_META_db_port" }}/postgres?sslmode=disable'
EOF
              destination = "secrets/db.env"
              env = true
            }

            template {
                data = file("includes/postgres-exporter.yaml")
                destination = "/local/config/postgres-exporter.yaml"
                change_mode = "noop"
            }

            config {
                image = "prometheuscommunity/postgres-exporter:latest"
                ports = ["postgres-exporter"]
            }

            resources {
                cpu = "${var.postgres_exporter_cpu}"
                memory = "${var.postgres_exporter_memory}"
            }
        }


    }
}
