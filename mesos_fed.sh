#!/bin/bash

ls
x="$(date)"
git clone https://git-wip-us.apache.org/repos/asf/mesos.git
sudo apt-get update
sudo apt-get install -y tar wget git
sudo apt-get install -y openjdk-7-jdk
sudo apt-get install -y autoconf libtool
sudo apt-get -y install build-essential python-dev python-boto libcurl4-nss-dev libsasl2-dev libsasl2-modules maven libapr1-dev libsvn-dev
echo "going under mesos dir..."
sleep 1m
cd mesos
echo "running bootstrap..."
sleep 1m
./bootstrap
mkdir build
cd build
../configure
echo "going to run make..."
sudo make
echo "going to run make check..."
sudo make check
echo "going to run make install..."
sudo make install
date
echo "$x"
#including FedModules

#cd ../../
#git clone https://github.com/huawei-cloudfederation/FedModules.git
#cd FedModules
#./build.sh
