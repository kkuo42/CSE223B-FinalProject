# #!/bin/bash

# This test emulates the behavior of vim

# The results should be ab on all servers

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist data/from0/a
assertExist data/from1/a
assertNotExist data/from2/a

echo "cat data/to2/a"
cat data/to2/a > /dev/null
sleep 1
assertFile data/from0/a "a"
assertFile data/from1/a "a"
assertFile data/from2/a "a"

echo "mv data/to2/a data/to2/a~"
mv data/to2/a data/to2/a~ > /dev/null
sleep 1
assertExist data/from0/a~
assertNotExist data/from0/a
assertExist data/from1/a~
assertNotExist data/from1/a
assertExist data/from2/a~
assertNotExist data/from2/a
assertExist data/to0/a~
assertNotExist data/to0/a
assertExist data/to1/a~
assertNotExist data/to1/a
assertExist data/to2/a~
assertNotExist data/to2/a

echo "writing ab to data/to1/a"
echo -n "ab" > data/to1/a
sleep 1
assertFile data/from0/a "ab"
assertFile data/from1/a "ab"
assertFile data/from2/a "ab"
assertFile data/to0/a "ab"
assertFile data/to1/a "ab"
assertFile data/to2/a "ab"

stop_jobs
pass_test
