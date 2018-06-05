#!/bin/bash

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "creating files a, b, c on servers 0, 1, 2"
echo -n "a" > data/to0/a
echo -n "b" > data/to1/b
echo -n "c" > data/to2/c
sleep 1

# verify files replicated
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"

assertNotExist "data/from0/b"
assertExist "data/from1/b"
assertExist "data/from2/b"

assertExist "data/from0/c"
assertNotExist "data/from1/c"
assertExist "data/from2/c"

# check zkdata
assertZkMetaEqual get /data/a a_data_init
assertZkMetaEqual get /data/b b_data_init
assertZkMetaEqual get /data/c c_data_init

echo "doing 5 reads on a from serv 0, 3 from serv 1, 1 from serv 2"
for ((n=0;n<5;n++)); do cat data/to0/a > /dev/null; done
for ((n=0;n<3;n++)); do cat data/to1/a > /dev/null; done
for ((n=0;n<1;n++)); do cat data/to2/a > /dev/null; done
assertExist "data/from2/a"
assertZkMetaEqual get /data/a a_data_fr

echo "doing 10 reads on b from serv 1, 3 from serv 0, 5 from serv 2"
for ((n=0;n<3;n++)); do cat data/to0/b > /dev/null; done
for ((n=0;n<10;n++)); do cat data/to1/b > /dev/null; done
for ((n=0;n<5;n++)); do cat data/to2/b > /dev/null; done
assertExist "data/from0/b"
assertZkMetaEqual get /data/b b_data_fr

echo "doing 50 reads on c from serv 0, 3 from serv 1, 10 from serv 2"
for ((n=0;n<50;n++)); do cat data/to0/c > /dev/null; done
for ((n=0;n<3;n++)); do cat data/to1/c > /dev/null; done
for ((n=0;n<10;n++)); do cat data/to2/c > /dev/null; done
assertExist "data/from1/c"
assertZkMetaEqual get /data/c c_data_fr

echo "doing 5 writes on a from serv 0, 10 from serv 1, 2 from serv 2"
for ((n=0;n<5;n++)); do echo "serv 0" >> data/to0/a; done
for ((n=0;n<10;n++)); do echo "serv 1" >> data/to1/a; done
for ((n=0;n<2;n++)); do echo "serv 2" >> data/to2/a; done
# TODO check metadata and all files have the correct contents
assertZkMetaEqual get /data/a a_data_fw

echo "doing 3 writes on b from serv 0, 10 from serv 1, 5 from serv 2"
for ((n=0;n<3;n++)); do echo "serv 0" >> data/to0/b; done
for ((n=0;n<10;n++)); do echo "serv 1" >> data/to1/b; done
for ((n=0;n<5;n++)); do echo "serv 2" >> data/to2/b; done
assertZkMetaEqual get /data/a a_data_fw

echo "doing 10 writes on c from serv 0, 1 from serv 1, 7 from serv 2"
for ((n=0;n<10;n++)); do echo "serv 0" >> data/to0/c; done
for ((n=0;n<1;n++)); do echo "serv 1" >> data/to1/c; done
for ((n=0;n<7;n++)); do echo "serv 2" >> data/to2/c; done
assertZkMetaEqual get /data/a a_data_fw

stop_jobs
pass_test
