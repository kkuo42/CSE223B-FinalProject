# #!/bin/bash

# This test creates a file from server a, reads it from server b, and then
# writes to it from server b.

# This test creates a file from server a, and then appends to it from server b.

# The results should be ab on all servers

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

# write initial file, check on server0, 1 not on 2
echo "writing a to data/to0/a"
echo -n "a" > data/to0/a
sleep 1
assertExist data/from0/a
assertExist data/from1/a
assertNotExist data/from2/a

# read file on server 1
echo "cat data/to1/a"
cat data/to1/a > /dev/null
sleep 1
assertFile data/from0/a "a"
assertFile data/from1/a "a"
assertFile data/to0/a "a"
assertFile data/to1/a "a"
assertNotExist data/from2/a

# append to file from replica 
echo "appending b to data/to1/a"
echo -n "b" >> data/to1/a
sleep 1
assertFile data/from0/a "ab"
assertFile data/from1/a "ab"
assertFile data/to0/a "ab"
assertFile data/to1/a "ab"
assertNotExist data/from2/a

# append to file from non replica
echo "appending c to data/to2/a"
echo -n "c" >> data/to2/a
sleep 1
assertFile data/from0/a "abc"
assertFile data/from1/a "abc"
assertFile data/from2/a "abc"
assertFile data/to0/a "abc"
assertFile data/to1/a "abc"
assertFile data/to2/a "abc"

# create new file on server 1
echo "writing c to data/to1/c"
echo -n "c" > data/to1/c
sleep 1
assertNotExist data/from0/c
assertExist data/from1/c
assertExist data/from2/c

# append to it from server 2
echo "appending d to data/to2/c"
echo -n "d" >> data/to2/c
sleep 1
assertNotExist data/from0/c
assertFile data/from1/c "cd"
assertFile data/from2/c "cd"
assertFile data/to1/c "cd"
assertFile data/to2/c "cd"
echo

# append to it from server 0
echo "appending e to data/to0/c"
echo -n "e" >> data/to0/c
sleep 1
assertFile data/from0/c "cde"
assertFile data/from1/c "cde"
assertFile data/from2/c "cde"
assertFile data/to0/c "cde"
assertFile data/to1/c "cde"
assertFile data/to2/c "cde"
echo

stop_jobs
pass_test
