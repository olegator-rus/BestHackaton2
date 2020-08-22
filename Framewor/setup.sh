#!/bin/bash

echo "installing micro text editor"
curl https://getmic.ro | bash

# Kibana Setup: https://www.elastic.co/guide/en/kibana/current/deb.html

echo "installing essential tools"
apt install gnupg git vim net-tools curl jq bat

echo "installing libs for running netcap"
apt install automake pkg-config autoconf
apt-get install -y apt-transport-https curl lsb-release wget autogen autoconf libtool gcc libpcap-dev linux-headers-generic git vim
echo "deb https://dl.bintray.com/wand/general $(lsb_release -sc) main" | tee -a /etc/apt/sources.list.d/wand.list
echo "deb https://dl.bintray.com/wand/libtrace $(lsb_release -sc) main" | tee -a /etc/apt/sources.list.d/wand.list
echo "deb https://dl.bintray.com/wand/libflowmanager $(lsb_release -sc) main" | tee -a /etc/apt/sources.list.d/wand.list
echo "deb https://dl.bintray.com/wand/libprotoident $(lsb_release -sc) main" | tee -a /etc/apt/sources.list.d/wand.list
curl --silent "https://bintray.com/user/downloadSubjectPublicKey?username=wand" | apt-key add -
apt-get update
apt install -y liblinear-dev libprotoident libprotoident-dev libprotoident-tools libtrace4-dev libtrace4-tools

wget https://github.com/ntop/nDPI/archive/3.0.tar.gz
tar xfz 3.0.tar.gz
cd nDPI-3.0 && ./autogen.sh && ./configure && make && make install

echo "importing elastic PGP key"
wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo apt-key add -

echo "installing apt-transport-https"
apt-get install apt-transport-https

echo "saving repository definition"
echo "deb https://artifacts.elastic.co/packages/7.x/apt stable main" | sudo tee -a /etc/apt/sources.list.d/elastic-7.x.list

apt-get update && apt-get install kibana

# Elastic Setup: https://www.elastic.co/guide/en/elasticsearch/reference/current/deb.html

apt-get update && apt-get install elasticsearch

# X-Pack Setup (installed by default): https://www.elastic.co/guide/en/x-pack/6.2/installing-xpack.html#xpack-package-installation

echo "adding elastic binaries to search path"
echo 'export PATH="$PATH:/usr/share/elasticsearch/bin/:/usr/share/kibana/bin"' >> ~/.bashrc
. ~/.bashrc

# TLS Setup: https://www.elastic.co/guide/en/kibana/current/configuring-tls.html

echo "installing unzip and xclip for micros external clipboard integration"
apt install unzip xclip

echo "update locate cache"
updatedb

mv kibana-server /usr/share/kibana/cert
chown -R kibana /usr/share/kibana/cert

# update elastic config to use data path on high volume storage medium:

elasticsearch.yml:

	data.path: /mnt/storage/elasticsearch

Then move data and update permissions:

	cp -r /var/lib/elasticsearch /mnt/storage/elasticsearch
	chown -R elasticsearch /mnt/storage/elasticsearch
