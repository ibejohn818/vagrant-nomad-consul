#!/usr/bin/env python

import ipaddress


BASE_IP = '192.168.60.0'

NOMAD_CONSUL = 3
APP = 3
DATA = 2


def server_map():
    sub = ipaddress.ip_address(BASE_IP) + 10

    servers = {
        "nomad": {}
        "consul": {}
        "app": {}
    }

    for i in range(NOMAD_CONSUL):
        key = i + 1
        




def main():
    server_map()


if __name__ == '__main__':
    main()
