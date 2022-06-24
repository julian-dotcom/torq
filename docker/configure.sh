#!/usr/bin/env bash

echo Configuring docker-compose file

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

cp example-docker-compose.yml docker-compose.yml
# https://stackoverflow.com/questions/16745988/sed-command-with-i-option-in-place-editing-works-fine-on-ubuntu-but-not-mac
sed -i.bak "s/<YourDBPassword>/$DBPASSWORD/" docker-compose.yml && rm docker-compose.yml.bak
sed -i.bak "s/<YourUIPassword>/$UIPASSWORD/" docker-compose.yml && rm docker-compose.yml.bak

echo 'Docker compose file (docker-compose.yml) created'
