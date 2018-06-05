# #!/bin/bash

# this test calls rm on a file from different servers in various scenarios.

# The results should be empty directories in each scenario

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "rm data/to0/a"
rm data/to0/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/from2/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
assertNotExist "data/to2/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "rm data/to1/a"
rm data/to1/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/from2/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
assertNotExist "data/to2/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "cat data/to1/a"
cat data/to1/a > /dev/null
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "rm data/to0/a"
rm data/to0/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/from2/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
assertNotExist "data/to2/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertNotExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "cat data/to2/a"
cat data/to2/a > /dev/null
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertExist "data/from2/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
assertExist "data/to2/a"
echo "rm data/to2/a"
rm data/to2/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/from2/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
assertNotExist "data/to2/a"
echo

stop_jobs
pass_test
