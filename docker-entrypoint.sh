#!/bin/sh

export CONTAINER_IP="$(awk 'END{print $1}' /etc/hosts)"

echo "Container IP: $CONTAINER_IP"
echo "Starting service ..."

./main