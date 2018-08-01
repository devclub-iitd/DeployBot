#!/bin/bash

# TODO:
# * Check errors from each command
# * Allow composition of yaml files

REPO=$1
MACHINE=$2


git clone --depth=1 $REPO
cd $REPO
docker-compose build
docker-compose push

eval "$(docker-machine env $MACHINE)"
docker-compose pull
docker-compose up --no-build
