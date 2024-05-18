# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'pp'
# require 'FileUtils'

Shared = Struct.new(:host, :vm)

g_bridge = "Intel(R) I211 Gigabit Network Connection"

# each server hash will use defaults if not present in its own hash
defaults = { 
  :hostname => "N/A",
  :ram => 4096,
  :cpus => 4,
  :box => "generic/debian12",
  # :shared => [add list of dedicated Shared dirs]
}

shared_global = [
  Shared.new("./shared/global", "/vagrant/global")
]

# this hostname should be last in the array of server configs,
# ansible provisioner will run on this hostname on all hosts vs sequentially
last_hostname = "data02"

# nomad_consul will run nomad & consul server on the same machines to lower box count
server_clusters = { 
  :nomad => [
    {
      :hostname => "nomad01",
      :ip => "192.168.60.11",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "nomad02",
      :ip => "192.168.60.12",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "nomad03",
      :ip => "192.168.60.13",
      :ram => 1024,
      :cpus => 2,
    },
  ],
  :consul => [
    {
      :hostname => "consul01",
      :ip => "192.168.60.14",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "consul02",
      :ip => "192.168.60.15",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "consul03",
      :ip => "192.168.60.16",
      :ram => 1024,
      :cpus => 2,
    },
  ],
  :nomad_consul => [
    {
      :hostname => "nomad_consul01",
      :ip => "192.168.60.14",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "nomad_consul02",
      :ip => "192.168.60.15",
      :ram => 1024,
      :cpus => 2,
    },
    {
      :hostname => "nomad_consul03",
      :ip => "192.168.60.16",
      :ram => 1024,
      :cpus => 2,
    },
  ]
}

client_nodes = [
  { 
    :hostname => "app01",
    :ip => "192.168.60.21",
    :box => "generic/debian12",
    :gui => false,
          },
  { 
    :hostname => "app02",
    :ip => "192.168.60.22",
    :box => "generic/debian12",
    :gui => false,
          },
  { 
    :hostname => "app03",
    :ip => "192.168.60.23",
    :box => "generic/debian12",
    :gui => false,
          },
  { 
    :hostname => "data01",
    :ip => "192.168.60.31",
    :box => "generic/debian12",
    :gui => false,
          },
  { 
    :hostname => "data02",
    :ip => "192.168.60.32",
    :box => "generic/debian12",
    :gui => false,
  }
]


# create the list of servers

# separate nomad & consul servers
servers = server_clusters[:nomad] + server_clusters[:consul]

# nomad & consul installed same server
# servers = server_clusters[:nomad_consul]

# merge clients
servers.concat(client_nodes)

# merge in defaults to all configs
servers.each_with_index do |s, k| 
  s.merge!(defaults) { |k, v, d| v }
  # set default dedicated path if not present
  if !s.key?(:shared) 
    s[:shared] =[
      Shared.new(
        "./shared/nodes/" + s[:hostname],
        "/vagrant/data"
      )
    ]
  end
end

# pp servers
#
# exit(0)

Vagrant.configure("2") do |config|
  # A common shared folder
  config.vm.synced_folder "./data/common", "/vagrant/common"

  config.vm.provision "shell", inline: <<-SHELL
    # echo "Global Provisioning goes here..."
  SHELL
  servers.each do |machine|
    config.vm.define machine[:hostname] do |node|
      node.vm.box = machine[:box]
      # node.vm.base_address = machine[:ip_addr]
      node.vm.network :private_network, ip: machine[:ip]
      node.vm.provision "shell", inline: <<-SHELL
        # Fix hostname
        echo "#{machine[:hostname]}" > /etc/hostname
      SHELL
      
      # invidiual shared folder
      # node.vm.synced_folder machine[:shared_folder_local], machine[:shared_folder_remote]
      # node.vm.synced_folder "./ansible", "/vagrant/ansible"


      machine[:shared].each do |share|
        FileUtils.mkdir_p(share.host) unless File.exists?(share.host)
        node.vm.synced_folder share.host, share.vm
      end

      # node.vm.provision "ansible_local" do |ansible|
      #     #ansible.verbose = "vvvv"
      #     ansible.playbook = "vagrant.yml"
      #     ansible.tags = []
      #     ansible.provisioning_path = "/vagrant/ansible"
      #     #ansible.vault_password_file = "~/.ansible/.vaultpw"
      # end

    if machine[:hostname] == last_hostname
      node.vm.provision "ansible" do |a|
          a.playbook = "ansible/vagrant.yml"
          a.limit = "all"
          a.groups = {
            "nomad_consul_servers" => ["nomad_consul0[1:3]"],
            "nomad_server" => ["nomad0[1:3]"],
            "consul_server" => ["consul0[1:3]"],
            "app" => ["app0[1:3]"],
            "data" => ["data0[1:2]"],
            "app:vars" => {"nomad_client_node_class" => "app"},
            "data:vars" => {"nomad_client_node_class" => "data"}
          }
      end
    end

 
      node.vm.provider "vmware_desktop" do |v|
        v.nat_device = "vmnet2"
        v.memory = machine[:memory]
        v.cpus = machine[:cpu]
      end

      node.vm.provider "virtualbox" do |vb|
        vb.gui = machine[:gui]
        vb.memory = machine[:ram]
        vb.cpus = machine[:cpus]
      end
    end
  end
end
