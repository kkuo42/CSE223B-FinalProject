#!/bin/bash

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

sleep 10
echo "killing backend 2"
kill $server2PID
