# #!/bin/bash

# This test creates a file from server a, reads it from server b, and then
# writes to it from server b.

# This test creates a file from server a, and then appends to it from server b.

# The results should be ab on all servers

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist data/from0/a
assertNotExist data/from1/a

echo "cat data/to1/a"
cat data/to1/a > /dev/null
sleep 1
assertFile data/from0/a "a"
assertFile data/from1/a "a"
assertFile data/to0/a "a"
assertFile data/to1/a "a"

echo "appending b to data/to1/a"
echo -n "b" >> data/to1/a
sleep 1
assertFile data/from0/a "ab"
assertFile data/from1/a "ab"
assertFile data/to0/a "ab"
assertFile data/to1/a "ab"

echo "writing c to data/to0/c"
echo "c" > data/to0/c
sleep 1

echo "appending d to data/to1/c"
echo "d" >> data/to1/c
sleep 1
assertFile data/from0/c "cd"
assertFile data/from1/c "cd"
assertFile data/to0/c "cd"
assertFile data/to1/c "cd"
echo

stop_jobs
pass_test
