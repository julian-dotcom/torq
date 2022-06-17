#!/usr/bin/env bash

read -p "Are you wish to delete Torq including data? (y/n)" -n 1 -r
echo    # (optional) move to a new line
if [[ $REPLY =~ ^[Yy]$ ]]
then
    docker-compose down -v
fi
