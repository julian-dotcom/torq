#!/usr/bin/env bash

echo Configuring docker-compose and torq.conf files

printf "\n"
echo Please specify where you want to add the Torq help commands
read -p "Directory (default: ~/.torq): " TORQDIR
eval TORQDIR="${TORQDIR:=$HOME/.torq}"
echo $TORQDIR
printf "\n"

# Set web UI password
printf "\n"
stty -echo
read -p "Please set a web UI password: " UIPASSWORD

while [[ -z "$UIPASSWORD" ]]; do
  printf "\n"
  read -p "The password cannot be empty, please try again: " UIPASSWORD
done

stty echo
printf "\n"

# Set web UI port number

printf "\n"
echo Please choose a port number for the web UI.
echo NB! Umbrel users needs to use a different port than 8080. Try 8081.
read -p "Port number (default: 8080): " UI_PORT
eval UI_PORT="${UI_PORT:=8080}"

while [[ ! $UI_PORT =~ ^[0-9]+$ ]] || [[ $UI_PORT -lt 1 ]] || [[ $UI_PORT -gt 65535 ]]; do
    read -p "Invalid port number. Please enter a valid port number from 1 through 65535: " UI_PORT
done

# Set network type

printf "\n"
stty -echo
read -p "Please choose network type (host/docker): " NETWORK

while [[ "$NETWORK" != "host" ]] && [[ "$NETWORK" != "docker" ]]; do
  printf "\n"
  read -p "Please choose network type (host/docker): " NETWORK
done

stty echo
printf "\n"

mkdir -p $TORQDIR

[ -f docker-compose.yml ] && rm docker-compose.yml

curl --location --silent --output "${TORQDIR}/torq.conf"            https://raw.githubusercontent.com/lncapital/torq/main/docker/example-torq.conf
if [[ "$NETWORK" == "host" ]]; then
  curl --location --silent --output "${TORQDIR}/docker-compose.yml" https://raw.githubusercontent.com/lncapital/torq/main/docker/example-docker-compose-host-network.yml
fi
if [[ "$NETWORK" == "docker" ]]; then
  curl --location --silent --output "${TORQDIR}/docker-compose.yml" https://raw.githubusercontent.com/lncapital/torq/main/docker/example-docker-compose.yml
fi

# https://stackoverflow.com/questions/16745988/sed-command-with-i-option-in-place-editing-works-fine-on-ubuntu-but-not-mac
#torq.conf setup
sed -i.bak "s/<YourUIPassword>/$UIPASSWORD/g" $TORQDIR/torq.conf          && rm $TORQDIR/torq.conf.bak
sed -i.bak "s/<YourPort>/$UI_PORT/g"          $TORQDIR/torq.conf          && rm $TORQDIR/torq.conf.bak
#docker-compose.yml setup
sed -i.bak "s/<Path>/$TORQDIR\/torq.conf/g"   $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak
if [[ "$NETWORK" == "docker" ]]; then
  sed -i.bak "s/<YourPort>/$UI_PORT/g"        $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak
fi
#start-torq setup
sed -i.bak "s/<YourPort>/$UI_PORT/g"          $TORQDIR/start-torq         && rm $TORQDIR/start-torq.bak

echo 'Docker compose file (docker-compose.yml) created in '$TORQDIR
echo 'Torq configuration file (torq.conf) created in '$TORQDIR

printf "\n"
