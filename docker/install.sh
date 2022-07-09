#!/usr/bin/env bash


echo Configuring docker-compose file

printf "\n"
echo Please where you want to add the Torq help commands
stty echo
printf "Directory (default: ~/.torq):"
read TORQDIR
stty echo
eval TORQDIR="${TORQDIR:=$HOME/.torq}"
echo $TORQDIR
printf "\n"

printf "\n"
echo Please set a database password
stty -echo
printf "DB Password: "
read DBPASSWORD
stty echo
printf "\n"

printf "\n"
echo Please set a web ui password
stty -echo
printf "UI Password: "
read UIPASSWORD
stty echo
printf "\n"
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
sed -i.bak "s/<YourDBPassword>/$DBPASSWORD/" $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak
sed -i.bak "s/<YourUIPassword>/$UIPASSWORD/" $TORQDIR/docker-compose.yml && rm $TORQDIR/docker-compose.yml.bak

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

echo "\nTry it out! Start Torq now with:"
echo "${Green}${TORQDIR}/${START_COMMAND}${NC}"
echo "\n"
