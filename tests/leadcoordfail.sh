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
sleep 2
echo "killing backend 0"
kill $server0PID
sleep 5

assertZkMetaEqual get /alivemeta/localhost:9501 9501_f2
assertZkMetaEqual get /alivemeta/localhost:9502 9502_f2
assertZkMetaEqual get /alivemeta/localhost:9601 9601_f2
assertZkMetaEqual get /alivemeta/localhost:9602 9602_f2

assertExist data/from1/a
assertExist data/from1/b
assertExist data/from1/c
assertExist data/from2/a
assertExist data/from2/b
assertExist data/from2/c

assertZkMetaEqual get /data/a a_data_fp
assertZkMetaEqual get /data/b b_data_fp
assertZkMetaEqual get /data/c c_data_fp

# check that alivemeta no longer has the old nodes
assertZkMetaEqual ls /alivemeta alive_a2

# ensure that client 0 still works fine
echo -n "c2" > data/to0/c2
assertExist data/from2/c2
assertExist data/from2/c2
echo -n "c3" > data/to2/c3

echo "sleeping and will check for rebalance"
sleep 15

assertZkMetaEqual2 get /alivemeta/localhost:9501 9501_f3 9501_f32
assertZkMetaEqual2 get /alivemeta/localhost:9502 9502_f3 9502_f32
assertZkMetaEqual2 get /alivemeta/localhost:9601 9601_f3 9601_f32
assertZkMetaEqual2 get /alivemeta/localhost:9602 9602_f3 9602_f32

stop_jobs
pass_test
