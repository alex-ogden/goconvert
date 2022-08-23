#!/bin/bash

# General script variables
site_directory="$HOME/sites/goconvert/"
container_name="goconvert"

# My-server-specific vars
scripts_directory="$HOME/.scripts/container-scripts"
script_name="30-start-$container_name.sh"

# Pull newest changes
pull_and_build() {
	cd "$site_directory" || exit
	git pull
	docker build -t "$container_name" .
}

# Remove currently running container
clean_and_run() {
	docker rm -f "$container_name"
	bash "$scripts_directory/$script_name"
}

pull_and_build
clean_and_run
