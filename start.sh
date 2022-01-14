#!/bin/bash

if [ "$1" = "client" ]
then
  sudo mkdir -p /etc/spleen/
  sudo cp spleen-client.service /etc/systemd/system/
  sudo cp spleen-client /etc/spleen/
  sudo cp .client.json /etc/spleen/
  systemctl enable spleen-client
  systemctl start spleen-client
  systemctl status spleen-client -l
elif [ "$1" = "server" ]
then
  sudo mkdir -p /etc/spleen/
  sudo cp spleen-server.service /etc/systemd/system/
  sudo cp spleen-server /etc/spleen/
  sudo cp .server.json /etc/spleen/
  systemctl enable spleen-server
  systemctl start spleen-server
  systemctl status spleen-server -l
else
  echo "Invalid parameter"
fi

