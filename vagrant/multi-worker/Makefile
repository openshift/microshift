# Export shell defined to support Ubuntu
export SHELL := $(shell which bash)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

microshift:
	make -C ../.. DEBUG=true
	cp ../../microshift $(PROJECT_DIR)
.PHONY: microshift
	
.vagrant:
	vagrant up

up: microshift .vagrant
.PHONY: up

down:
	vagrant destroy -f
	rm .token_installed
.PHONY: down

install-worker-token: .vagrant
	vagrant ssh worker -- sudo mkdir -p /var/lib/microshift/resources/kubelet
	vagrant ssh master -- sudo cat /var/lib/microshift/resources/kubelet/bootstrap-kubeconfig | vagrant ssh worker -- sudo tee /var/lib/microshift/resources/kubelet/bootstrap-kubeconfig

.PHONY: install-worker-token

.token_installed: install-worker-token
	touch .token_installed

start-worker: .token_installed
	# we need a separate service definition
	vagrant ssh worker -- sudo sed -i \'s/run/run --roles=node/g\' /usr/lib/systemd/system/microshift.service
	vagrant ssh worker -- sudo systemctl enable --now microshift.service

restart-worker:
	vagrant ssh worker -- sudo systemctl restart  microshift.service

restart-master:
	vagrant ssh master -- sudo systemctl restart  microshift.service

.PHONY: start-worker restart-worker restart-master

update-worker: microshift .vagrant 
	vagrant upload microshift /vagrant/microshift worker
	vagrant ssh worker -- sudo rm /usr/bin/microshift
	vagrant ssh worker -- sudo mv /vagrant/microshift /usr/bin
	vagrant ssh worker -- sudo restorecon -F -v /usr/bin/microshift
	vagrant ssh worker -- sudo systemctl kill microshift || true
	vagrant ssh worker -- sudo systemctl kill microshift || true # fast and dirty... just the worker ;)
.PHONY: update-worker

update-master: microshift .vagrant
	vagrant upload microshift /vagrant/microshift master 
	vagrant ssh master -- sudo rm /usr/bin/microshift
	vagrant ssh master -- sudo mv /vagrant/microshift /usr/bin
	vagrant ssh master -- sudo restorecon -F -v /usr/bin/microshift
	vagrant ssh master -- sudo systemctl kill microshift || true 
	vagrant ssh master -- sudo systemctl kill microshift || true # fast and dirty.. just the master? ':D  
.PHONY: update-master

kubeconfig: .vagrant
	vagrant ssh master -- sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > kubeconfig

worker-logs-follow: .vagrant
	vagrant ssh worker -- sudo journalctl -o cat -f -u microshift.service

master-logs-follow: .vagrant
	vagrant ssh master -- sudo journalctl -o cat -f -u microshift.service

worker-logs: .vagrant
	vagrant ssh worker -- sudo journalctl -o cat -u microshift.service

master-logs: .vagrant
	vagrant ssh master -- sudo journalctl -o cat -u microshift.service

master-dlv: .vagrant
	vagrant ssh master -- /vagrant/runDlv.sh

.PHONY: worker-logs worker-logs-follow master-logs master-logs-follow 

