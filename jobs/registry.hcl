variable "host_placement" {
  type = string
  default = "app03"
}

job "registry" {
    datacenters = ["global"]
    region = "dc1"
    type = "service"

    constraint {
      attribute = "${node.class}"
      operator = "="
      value = "app"
    }

    constraint {
      attribute = "${attr.unique.hostname}"
      operator = "="
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

                "traefik.http.routers.dockreg.rule=Host(`registry.service.dc1.consul`)",
                "traefik.http.routers.dockreg.entrypoints=http",
                "traefik.http.routers.dockreg.middlewares=dockreg-https",
                "traefik.http.middlewares.dockreg-https.redirectscheme.scheme=https",

                "traefik.http.routers.dockreg-ssl.rule=Host(`registry.service.dc1.consul`)",
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
                data = file("config/registry.yml")
                destination = "local/registry/config.yml"
            }


            driver = "docker"

            env {
            }

            config {
                privileged=true
                image = "registry:2.4"
                ports = ["http"]
                volumes = [
                    "local/registry:/etc/docker/registry"
                ]
            }
            resources {
                memory=1024
                disk=1024
                cpu=2500
            }
        
        }

    
    }
}
