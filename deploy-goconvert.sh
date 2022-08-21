#!/bin/bash

# Vars
site_directory="$HOME/sites/goconvert/"
scripts_directory="$HOME/.scripts/container-scripts"
container_name="goconvert"
script_name="30-start-$container_name.sh"

# Pull newest changes
echo "LOG: Pulling latest changes"
cd "$site_directory" || exit
git pull

# Build new image
echo "LOG: Building new container image"
docker build -t "$container_name" .

# Remove currently running container
echo "LOG: Removing currently running container"
docker rm -f "$container_name"

# Deploy the new container image
echo "LOG: Deploying new container image"
bash "$scripts_directory/$script_name"
