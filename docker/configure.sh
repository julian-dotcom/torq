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
sed -i '' "s/\<YourDBPassword\>/$DBPASSWORD/" docker-compose.yml
sed -i '' "s/\<YourUIPassword\>/$UIPASSWORD/" docker-compose.yml

echo 'Docker compose file (docker-compose.yml) created'
