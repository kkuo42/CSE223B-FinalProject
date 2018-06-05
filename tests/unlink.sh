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
assertNotExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
echo "rm data/to0/a"
rm data/to0/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
assertNotExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
echo "rm data/to1/a"
rm data/to1/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
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
echo "rm data/to0/a"
rm data/to0/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
echo

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist "data/from0/a"
# It should exist here now because the primary/replica metadata remains on the
# keeper even though the file was marked as deleted. Write gets sent to the replica
assertExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
echo "cat data/to1/a"
cat data/to1/a > /dev/null
sleep 1
assertExist "data/from0/a"
assertExist "data/from1/a"
assertExist "data/to0/a"
assertExist "data/to1/a"
echo "rm data/to1/a"
rm data/to1/a > /dev/null
sleep 1
assertNotExist "data/from0/a"
assertNotExist "data/from1/a"
assertNotExist "data/to0/a"
assertNotExist "data/to1/a"
echo

stop_jobs
pass_test
