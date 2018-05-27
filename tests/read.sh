# #!/bin/bash

# this script creates a keeper locally. 
# it expects zookeeper to be in the same directory
# with the datadir in zookeeper-3.4.12/conf/zoo.cfg set to zkdata also in this directory

# the script lauches clients+servers as pairs so that the client is guaranteed to be assigned 
# to that server because the frontend will assign most recently joined server

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1

echo "cat data/to0/a"
cat data/to0/a > /dev/null
sleep 1
assertExist "data/from0/a"
assertNotExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"

echo "cat data/to1/a"
cat data/to1/a > /dev/null
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"

stop_jobs
pass_test
