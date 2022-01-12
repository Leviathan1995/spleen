#!/bin/bash

if [ "$1" = "client" ]
then
  mkdir -p /etc/spleen/
  cp spleen-client.service /etc/systemd/system/
  cp spleen-client /etc/spleen/
  cp .client.json /etc/spleen/
  systemctl enable spleen-client
  systemctl start spleen-client
  systemctl status spleen-client -l
elif [ "$1" = "server" ]
then
  mkdir -p /etc/spleen/
  cp spleen-server.service /etc/systemd/system/
  cp spleen-server /etc/spleen/
  cp .server.json /etc/spleen/
  systemctl enable spleen-server
  systemctl start spleen-server
  systemctl status spleen-server -l
else
  echo "Invalid parameter"
fi