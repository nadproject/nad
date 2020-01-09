# -*- mode: ruby -*-

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.synced_folder '.', '/vagrant/nad'
  config.vm.network "forwarded_port", guest: 3000, host: 3000
  config.vm.network "forwarded_port", guest: 8080, host: 8080
  config.vm.network "forwarded_port", guest: 5432, host: 5433

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "scripts/vagrant/configure.yml"
    ansible.extra_vars = {
      ansible_ssh_user: 'vagrant'
    }
  end

  config.vm.provider "virtualbox" do |v|
    v.memory = 4000
    v.cpus = 2
  end
end
