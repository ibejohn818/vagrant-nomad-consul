variable "node_exporter_cpu" {
  type = string
  description = "Nomad CPU Allocation"
  default = "100"
}

variable "node_exporter_memory" {
  type = string
  description = "Nomad Memory Allocation"
  default = "64"
}

job "node-exporter" {
    datacenters = ["dc1"]
    region = "global"
    type = "system"

    group "containers" {

        network {
            port "http" {}
        }

        service {
            name = "node-exporter"
            port = "http"
            tags = [ "node-exporter" ]
            check {
                name = "http-tcp"
                type = "tcp"
                interval = "5s"
                timeout = "2s"
            }
        }

        task "node-exporter" {
            driver = "docker"
            config {
                image = "prom/node-exporter:v1.3.1"
                args = [
                    "--path.procfs", "/host/proc",
                    "--path.sysfs", "/host/sys",
                    "--collector.filesystem.ignored-mount-points", "^/(sys|proc|dev|host|etc)($|/)",
                    "--web.listen-address=:${NOMAD_PORT_http}"
                ]
                
                network_mode = "host"

                volumes = [
                    "/proc:/host/proc",
                    "/sys:/host/sys",
                    "/:/rootfs",
                ]
                ulimit {
                    nproc = "1024"
                    nofile = "70000:70000"
                }
            }
            
            resources {
                cpu = "${var.node_exporter_cpu}"
                memory = "${var.node_exporter_memory}"
            }
        }

        restart {
            attempts = 3
            delay    = "15s"
            interval = "1m"
            mode     = "delay"
        }

    }
}
