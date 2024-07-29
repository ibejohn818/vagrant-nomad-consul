#!/usr/bin/env bash

vagrant ssh consul01 -c 'sudo systemctl stop consul && sudo rm -rf /var/lib/consul/* && sudo systemctl start consul'
vagrant ssh consul02 -c 'sudo systemctl stop consul && sudo rm -rf /var/lib/consul/* && sudo systemctl start consul'
vagrant ssh consul03 -c 'sudo systemctl stop consul && sudo rm -rf /var/lib/consul/* && sudo systemctl start consul'
