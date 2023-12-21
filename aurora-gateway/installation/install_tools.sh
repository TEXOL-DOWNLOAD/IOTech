#!/bin/bash

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo "curl is not installed, starting installation..."
    sudo apt-get update
    sudo apt-get install -y curl
    echo "curl installation completed!"
else
    echo "curl is already installed."
fi

# Check if json_pp is installed
if ! command -v json_pp &> /dev/null; then
    echo "json_pp is not installed, starting installation..."
    sudo apt-get update
    sudo apt-get install -y libjson-pp-perl
    echo "json_pp installation completed!"
else
    echo "json_pp is already installed."
fi
