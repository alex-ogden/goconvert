#!/bin/bash

# Vars
site_directory="$HOME/sites/goconvert/"
scripts_directory="$HOME/.scripts/container-scripts"
container_name="goconvert"
script_name="30-start-$container_name.sh"

# Build the new image with newest changes
cd "$site_directory" || exit
git pull
docker build -t "$container_name" .

# Remove currently running container
docker rm -f "$container_name"

# Deploy the new container image
bash "$scripts_directory/$script_name"
