#!/bin/bash

x="$(date)"
mkdir src
cd src
git clone https://git-wip-us.apache.org/repos/asf/mesos.git
sudo apt-get update
sudo apt-get install -y tar wget git
sudo apt-get install -y openjdk-7-jdk
sudo apt-get install -y autoconf libtool
sudo apt-get -y install build-essential python-dev python-boto libcurl4-nss-dev libsasl2-dev libsasl2-modules maven libapr1-dev libsvn-dev
cd mesos
./bootstrap
mkdir build
cd build
../configure
echo "going to run make..."
sudo make -j 8
echo "going to run make check..."
sudo make check -j 8
echo "going to run make install..."
sudo make install -j 8
date
echo "$x"
