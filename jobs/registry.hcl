variable "host_placement" {
  type = string
  default = ".*"
}

job "registry" {
    datacenters = ["global"]
    region = "dc1"
    type = "system"

    constraint {
      attribute = "${node.class}"
      operator = "="
      value = "app"
    }

    constraint {
      attribute = "${attr.unique.hostname}"
      operator = "regexp"
      value = "${var.host_placement}"
    }

    group "server" {
        network {
          port "http" {to =  5000}
        }

        service {


            name = "registry"
            port = "http"

            // check {
            //     type     = "tcp"
            //     port     = "http"
            //     interval = "25s"
            //     timeout  = "5s"
            // }

            tags = [
                "registry",
                "traefik.enable=true",

                "traefik.http.routers.dockreg.rule=Host(`registry.service.dc1.consul`) || Host(`registry.vagrant.local`)",
                "traefik.http.routers.dockreg.entrypoints=http",
                "traefik.http.routers.dockreg.middlewares=dockreg-https",
                "traefik.http.middlewares.dockreg-https.redirectscheme.scheme=https",

                "traefik.http.routers.dockreg-ssl.rule=Host(`registry.service.dc1.consul`) || Host(`registry.vagrant.local`)",
                "traefik.http.routers.dockreg-ssl.entrypoints=https",
                "traefik.http.routers.dockreg-ssl.tls=true",
              

            ]
            task = "registry"
            
        }

        task "registry" {
            # template { 
            #     data = file("jobs/templates/htpasswd")
            #     destination = "secrets/htpasswd"
            # }
            #
            template { 
                data = file("includes/registry.yml")
                destination = "local/registry/config.yml"
                change_mode = "noop"
            }


            driver = "docker"

            env {
            }

            config {
                privileged=true
                auth_soft_fail=true
                image = "registry:2.8"
                ports = ["http"]
                volumes = [
                    "/vagrant/common/registry:/var/lib/registry",
                    "local/registry:/etc/docker/registry"
                ]
            }
            resources {
                memory=1024
                cpu=2501
            }
        
        }

    
    }
}
