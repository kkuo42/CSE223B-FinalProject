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

# before killing make sure alivemeta has all nodes
assertZkMetaEqual ls /alivemeta alive_b1

echo "sleeping before killing"
sleep 1
echo "killing backend 2"
kill $server2PID

echo "server 2 killed, immediately issuing client2 write"
# ensure that client 2 still works fine
echo -n "c2" > data/to2/c2
assertExist data/from0/c2
assertExist data/from1/c2

# give just a sec to move data
sleep 1

assertZkMetaEqual2 get /alivemeta/localhost:9500 9500_ff1 9500_ff2
assertZkMetaEqual2 get /alivemeta/localhost:9501 9501_ff1 9501_ff2
assertZkMetaEqual2 get /alivemeta/localhost:9600 9600_ff1 9600_ff2
assertZkMetaEqual2 get /alivemeta/localhost:9601 9601_ff1 9601_ff2

assertExist data/from0/a
assertExist data/from0/b
assertExist data/from0/c
assertExist data/from1/a
assertExist data/from1/b
assertExist data/from1/c

assertZkMetaEqual get /data/a a_data_fb
assertZkMetaEqual get /data/b b_data_fb
assertZkMetaEqual get /data/c c_data_fb

# check that alivemeta no longer has the old nodes
assertZkMetaEqual ls /alivemeta alive_a1

stop_jobs
pass_test
