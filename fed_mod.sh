#!/bin/bash

cd src
git clone https://github.com/huawei-cloudfederation/FedModules.git
cd FedModules
./build.sh
export MESOS_HOME_DIR="/home/ubuntu/src/mesos"             #may need to update this later according the path of installation
cd ../mesos/build
sudo ./bin/mesos-master.sh --ip=127.0.0.1 --work_dir=$HOME --modules="file://../../FedModules/FedModules.json" --allocator="mesos_fed_allocator_module"
