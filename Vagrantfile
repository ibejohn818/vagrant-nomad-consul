# -*- mode: ruby -*-
# vi: set ft=ruby :

g_bridge = "Intel(R) I211 Gigabit Network Connection"

servers = [
  {
    :hostname => "server01",
    :ip => "192.168.60.11",
    :box => "generic/debian12",
    :ram => 1024,
    :cpus => 2,
    :gui => false,
    :shared_folder_local => "./data/server01",
    :shared_folder_remote => "/vagrant/data",
  },
  {
    :hostname => "server02",
    :ip => "192.168.60.12",
    :box => "generic/debian12",
    :ram => 1024,
    :cpus => 2,
    :gui => false,
    :shared_folder_local => "./data/server02",
    :shared_folder_remote => "/vagrant/data",
  },
  {
    :hostname => "server03",
    :ip => "192.168.60.13",
    :box => "generic/debian12",
    :ram => 1024,
    :cpus => 2,
    :gui => false,
    :shared_folder_local => "./data/server03",
    :shared_folder_remote => "/vagrant/data",
  },
  { 
    :hostname => "app01",
    :ip => "192.168.60.21",
    :box => "generic/debian12",
    :ram => 4096,
    :cpus => 4,
    :gui => false,
    :shared_folder_local => "./data/app01",
    :shared_folder_remote => "/vagrant/data",
  },
  { 
    :hostname => "app02",
    :ip => "192.168.60.22",
    :box => "generic/debian12",
    :ram => 4096,
    :cpus => 4,
    :gui => false,
    :shared_folder_local => "./data/app02",
    :shared_folder_remote => "/vagrant/data",
  },
  { 
    :hostname => "app03",
    :ip => "192.168.60.23",
    :box => "generic/debian12",
    :ram => 4096,
    :cpus => 4,
    :gui => false,
    :shared_folder_local => "./data/app03",
    :shared_folder_remote => "/vagrant/data",
  },
  { 
    :hostname => "data01",
    :ip => "192.168.60.31",
    :box => "generic/debian12",
    :ram => 4096,
    :cpus => 4,
    :gui => false,
    :shared_folder_local => "./data/data01",
    :shared_folder_remote => "/vagrant/data",
  },
  { 
    :hostname => "data02",
    :ip => "192.168.60.32",
    :box => "generic/debian12",
    :ram => 4096,
    :cpus => 4,
    :gui => false,
    :shared_folder_local => "./data/data02",
    :shared_folder_remote => "/vagrant/data",
  }
]



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
      node.vm.synced_folder machine[:shared_folder_local], machine[:shared_folder_remote]
      node.vm.synced_folder "./ansible", "/vagrant/ansible"

      # node.vm.provision "ansible_local" do |ansible|
      #     #ansible.verbose = "vvvv"
      #     ansible.playbook = "vagrant.yml"
      #     ansible.tags = []
      #     ansible.provisioning_path = "/vagrant/ansible"
      #     #ansible.vault_password_file = "~/.ansible/.vaultpw"
      # end

    if machine[:hostname] == "data02"
      node.vm.provision "ansible" do |a|
          a.playbook = "ansible/vagrant.yml"
          a.limit = "all"
          a.groups = {
            "nomad_consul_servers" => ["server0[1:3]"],
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
