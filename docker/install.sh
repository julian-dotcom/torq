#!/usr/bin/env bash

echo Configuring docker-compose file

printf "\n"
echo Please where you want to add the Torq help commands
stty echo
printf "Directory (default: ~/.torq):"
read TORQDIR
stty echo
TORQDIR="${TORQDIR:=$HOME/.torq}"
mkdir -p $TORQDIR
printf "\n"
echo $TORQDIR

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

[ -f docker-compose.yml ] && rm docker-compose.yml

curl --location --silent --output $TORQDIR"/docker-compose.yml" https://raw.githubusercontent.com/lncapital/torq/main/docker/example-docker-compose.yml

curl --location --silent --output $TORQDIR"/start.sh" https://raw.githubusercontent.com/lncapital/torq/main/docker/start.sh
curl --location --silent --output $TORQDIR"/stop.sh" https://raw.githubusercontent.com/lncapital/torq/main/docker/stop.sh
curl --location --silent --output $TORQDIR"/delete.sh" https://raw.githubusercontent.com/lncapital/torq/main/docker/delete.sh
curl --location --silent --output $TORQDIR"/configure.sh" https://raw.githubusercontent.com/lncapital/torq/main/docker/configure.sh

# https://stackoverflow.com/questions/16745988/sed-command-with-i-option-in-place-editing-works-fine-on-ubuntu-but-not-mac
sed -i.bak "s/<YourDBPassword>/$DBPASSWORD/"  $TORQDIR"/docker-compose.yml" && rm $TORQDIR"/docker-compose.yml.bak"
sed -i.bak "s/<YourUIPassword>/$UIPASSWORD/" $TORQDIR"/docker-compose.yml" && rm $TORQDIR"/docker-compose.yml.bak"

echo 'Docker compose file (docker-compose.yml) created in '$TORQDIR
echo 'Start Torq with:\n\n'
echo 'sh '$TORQDIR'/start.sh'
