#!/bin/sh

set -e

DEST_BIN="/usr/local/bin/ward"
DEST_CONF="/etc/ward/config.yaml"
DEST_INITD="/etc/init.d/ward"

echo "building..."
go build -o ward ./cmd/ward/

echo "installing binary..."
install -m 755 ward "${DEST_BIN}"

echo "installing config..."
mkdir -p /etc/ward
if [ ! -f "${DEST_CONF}" ]; then
    install -m 644 config.yaml "${DEST_CONF}"
    echo "  config installed to ${DEST_CONF}"
else
    echo "  config already exists, skipping"
fi

echo "installing init script..."
install -m 755 ward.initd "${DEST_INITD}"

echo "done. use:"
echo "  rc-service ward start"
echo "  rc-update add ward default"
