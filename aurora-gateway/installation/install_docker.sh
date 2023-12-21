#!/bin/bash

cpu_arch=$(uname -m)
legal_arch=("amd64" "x86_64" "aarch64")
distro_id=$(cat /etc/os-release | grep -i '^ID=' | awk -F'=' '{print tolower($2)}')
distro_ver=$(cat /etc/os-release | grep -i '^VERSION_ID=' | awk -F'=' '{print tolower($2)}' | sed 's/"//g')


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

# Rest of the script...
distro_codename=$(lsb_release -cs)
containerd_version="1.6.9-1"
docker_ce_cli_version="24.0.7-1"
docker_ce_version="24.0.7-1"
docker_compose_plugin_version="2.21.0-1"


containerd=containerd.io_"$containerd_version"_"$arch".deb
docker_ce_cli=docker-ce-cli_"$docker_ce_cli_version"~"$distro_id"."$distro_ver"~"$distro_codename"_"$arch".deb
docker_ce=docker-ce_"$docker_ce_version"~"$distro_id"."$distro_ver"~"$distro_codename"_"$arch".deb
docker_compose=docker-compose-plugin_"$docker_compose_plugin_version"~"$distro_id"."$distro_ver"~"$distro_codename"_"$arch".deb


wget https://download.docker.com/linux/"$distro_id"/dists/"$distro_codename"/pool/stable/"$arch"/"$containerd"
wget https://download.docker.com/linux/"$distro_id"/dists/"$distro_codename"/pool/stable/"$arch"/"$docker_ce_cli"
wget https://download.docker.com/linux/"$distro_id"/dists/"$distro_codename"/pool/stable/"$arch"/"$docker_ce"
wget https://download.docker.com/linux/"$distro_id"/dists/"$distro_codename"/pool/stable/"$arch"/"$docker_compose"

# Install Docker Engine
sudo dpkg -i $docker_ce_cli
sudo dpkg -i $containerd
sudo dpkg -i $docker_ce
# Install Docker-Compose
sudo dpkg -i $docker_compose

rm $docker_ce_cli $containerd $docker_ce $docker_compose

# Join Docker group
sudo groupadd docker
echo 'current user: '$USER
sudo usermod -aG docker $USER
