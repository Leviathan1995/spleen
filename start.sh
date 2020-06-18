#!/bin/bash

if [ "$1" = "client" ]
then
  sudo mkdir -p /root/proxy/
  sudo cp spleen-client.service /etc/systemd/system/
  sudo cp spleen-client /root/proxy/
  sudo cp .client.json /root/proxy/
  systemctl enable spleen-client
  systemctl start spleen-client
  systemctl status spleen-client -l
elif [ "$1" = "server" ]
then
  sudo mkdir -p /root/proxy/
  sudo cp spleen-server.service /etc/systemd/system/
  sudo cp spleen-server /root/proxy/
  sudo cp .server.json /root/proxy/
  systemctl enable spleen-server
  systemctl start spleen-server
  systemctl status spleen-server -l
else
  echo "Invalid parameter"
fi