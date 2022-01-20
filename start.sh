#!/bin/bash

if [ "$1" = "client" ]
then
  sudo mkdir -p /etc/spleen/
  sudo cp spleen-client.service /etc/systemd/system/
  sudo cp spleen-client /etc/spleen/
  sudo cp .spleen-client.json /etc/spleen/
  sudo systemctl enable spleen-client
  sudo systemctl start spleen-client
  sudo systemctl status spleen-client -l
elif [ "$1" = "server" ]
then
  sudo mkdir -p /etc/spleen/
  sudo cp spleen-server.service /etc/systemd/system/
  sudo cp spleen-server /etc/spleen/
  sudo cp .spleen-server.json /etc/spleen/
  sudo systemctl enable spleen-server
  sudo systemctl start spleen-server
  sudo systemctl status spleen-server -l
else
  echo "Invalid parameter"
fi

