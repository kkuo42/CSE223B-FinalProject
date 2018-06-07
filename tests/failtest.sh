#!/bin/bash

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "writing data to clients"
echo -n "a" > data/to0/a
echo -n "b" > data/to1/b
echo -n "c" > data/to2/c
echo "have to1 issue more writes to c so it gets it"
echo "1" >> data/to1/c
echo "1" >> data/to1/c

echo "sleeping before killing"
sleep 2
echo "killing backend 2"
kill $server2PID
sleep 5

# TODO ensure that client 2 still works fine
# TODO ensure that metadata is moved properly/everything assigned
assertZkMetaEqual get /alivemeta/localhost:9500 9500_f1
assertZkMetaEqual get /alivemeta/localhost:9501 9501_f1
assertZkMetaEqual get /alivemeta/localhost:9600 9600_f1
assertZkMetaEqual get /alivemeta/localhost:9601 9601_f1

stop_jobs
pass_test
