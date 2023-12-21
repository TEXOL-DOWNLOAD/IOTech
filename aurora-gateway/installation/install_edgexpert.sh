#!/bin/bash

xpert_version="2.3.7"

export LICENSE_PATH=$1

cpu_arch=$(uname -m)
legal_arch=("amd64" "x86_64" "aarch64")

# Check if cpu_arch is in legal_arch
if [[ ! " ${legal_arch[@]} " =~ " ${cpu_arch} " ]]; then
  echo "Architecture not supported."
  exit 1
fi

# Set arch based on cpu_arch
if [[ $cpu_arch == "amd64" || $cpu_arch == "x86_64" ]]; then
  arch="amd64"
elif [[ $cpu_arch == "aarch64" ]]; then
  arch="arm64"
else
  echo "Architecture not supported."
  exit 1
fi

edgexpert_deb=edgexpert-"$xpert_version"_"$arch".deb

# Download EdgeXpert CLI installation
wget https://iotech.jfrog.io/artifactory/edgexpert-releases/edgexpert/deb/edgexpert-"$edgexpert_deb"

sudo dpkg -i "$edgexpert_deb"
rm -rf "$edgexpert_deb"

# Install EdgeXpert License
if [ -n "$LICENSE_PATH" ]; then
  echo "Install edgexpert license: $LICENSE_PATH"
  edgexpert license install $LICENSE_PATH
  edgexpert license check
fi