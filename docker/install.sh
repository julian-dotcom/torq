#!/usr/bin/env bash

echo Configuring docker-compose file

# Set Torq help commands directory

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

printf "\n"

mkdir -p $TORQDIR

[ -f docker-compose.yml ] && rm docker-compose.yml

START_COMMAND='start-torq'
STOP_COMMAND='stop-torq'
UPDATE_COMMAND='update-torq'
DELETE_COMMAND='delete-torq'

curl --location --silent --output "${TORQDIR}/docker-compose.yml" https://raw.githubusercontent.com/lncapital/torq/main/docker/example-docker-compose.yml
curl --location --silent --output "${TORQDIR}/${START_COMMAND}" https://raw.githubusercontent.com/lncapital/torq/main/docker/start.sh
curl --location --silent --output "${TORQDIR}/${STOP_COMMAND}" https://raw.githubusercontent.com/lncapital/torq/main/docker/stop.sh
curl --location --silent --output "${TORQDIR}/${UPDATE_COMMAND}" https://raw.githubusercontent.com/lncapital/torq/main/docker/update.sh
curl --location --silent --output "${TORQDIR}/${DELETE_COMMAND}" https://raw.githubusercontent.com/lncapital/torq/main/docker/delete.sh

chmod +x $TORQDIR/$START_COMMAND
chmod +x $TORQDIR/$STOP_COMMAND
chmod +x $TORQDIR/$UPDATE_COMMAND
chmod +x $TORQDIR/$DELETE_COMMAND

# https://stackoverflow.com/questions/16745988/sed-command-with-i-option-in-place-editing-works-fine-on-ubuntu-but-not-mac
sed -i.bak "s/<YourUIPassword>/$UIPASSWORD/" $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak
sed -i.bak "s/<YourPort>/$UI_PORT/g" $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak
sed -i.bak "s/<YourPort>/$UI_PORT/g" $TORQDIR/start-torq && rm $TORQDIR/start-torq.bak

echo 'Docker compose file (docker-compose.yml) created in '$TORQDIR

Green='\033[0;32m' # Green text color
Cyan='\033[0;36m'
Red='\033[0;31m'
NC='\033[0m' ## Reset text color


echo "We have added these scripts to ${TORQDIR}:\n"
echo "${Cyan}${START_COMMAND}${NC}\t (This command starts Torq)"
echo "${Cyan}${STOP_COMMAND}${NC}\t (This command stops Torq)"
echo "${Cyan}${UPDATE_COMMAND}${NC}\t (This command updates Torq)"
echo "${Red}${DELETE_COMMAND}${NC}\t (WARNING: This command deletes Torq _including_ all collected data!)"


echo "${Green}Optional:${NC} you can add these scripts to your PATH by running:"
echo "sudo ln -s ${TORQDIR}/* /usr/local/bin/"

echo "\nTry it out now! Make sure the Docker daemon is running, and then start Torq with:"
echo "${Green}${TORQDIR}/${START_COMMAND}${NC}"
echo "\n"
