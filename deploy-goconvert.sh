#!/bin/bash

site_directory="/root/sites/goconvert/"
scripts_directory="/root/.scripts/container-scripts/"
container_name="goconvert"
script_name="30-start-$container_name.sh"

cd "$site_directory"
git pull
docker build -t "$container_name" .
cd "$scripts_directory"
docker rm -f "$container_name"
bash "$script_name"
clear
docker logs --follow "$container_name"
