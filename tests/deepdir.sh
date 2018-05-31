#!/bin/bash

# deep dir means nested directories. 


source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers


# read a file in the deep dir of another server
echo
echo "read a file in the deep dir of another server"

echo "creating nested directories data/to0/a/b/c"
mkdir -p data/to0/a/b/c

assertExist 	"data/from0/a/b/c" # check that nesting works
assertExist 	"data/to1/a/b/c"

echo "writing something to data/to0/a/b/c/something"
echo "something" >> data/to0/a/b/c/something
sleep 1

echo "cat data/to0/a/b/c/something"
cat data/to0/a/b/c/something > /dev/null
sleep 1
assertExist 	"data/from0/a/b/c/something"
assertNotExist 	"data/from1/a/b/c/something"
assertExist 	"data/to0/a/b/c/something"
assertExist 	"data/to1/a/b/c/something"

echo "cat data/to0/a/b/c/something"
cat data/to1/a/b/c/something > /dev/null
sleep 1
assertExist "data/from0/a/b/c/something"
assertExist "data/from1/a/b/c/something"
assertExist "data/to0/a/b/c/something"
assertExist "data/to1/a/b/c/something"

assertFilesEqual "data/to0/a/b/c/something" "data/to1/a/b/c/something"


# write a file in the deep dir of another server
echo
echo "write a file in the deep dir of another server"

echo "creating nested directories data/to0/e/f/g"
mkdir -p data/to0/e/f/g

echo "writing something to data/to0/e/f/g/something"
echo "something" >> data/to0/e/f/g/something
sleep 1

echo "writing something else to data/to1/e/f/g/something"
echo "something else" >> data/to1/e/f/g/something
sleep 1

assertExist "data/from0/e/f/g/something"
assertExist "data/from1/e/f/g/something"
assertExist "data/to0/e/f/g/something"
assertExist "data/to1/e/f/g/something"

touch temp
echo "something" >> temp
echo "something else" >> temp
assertFilesEqual "data/to0/e/f/g/something" "temp"
assertFilesEqual "data/to1/e/f/g/something" "temp"
rm temp


# create a file in the deep dir of another server
echo 
echo create a file in the deep dir of another server

echo "creating nested directories data/to0/h/i/j"
mkdir -p data/to0/h/i/j

echo "writing something else to data/to1/h/i/j/something"
echo "something" >> data/to1/h/i/j/something
sleep 1

assertNotExist 	"data/from0/h/i/j/something"
assertExist 	"data/from1/h/i/j/something"
assertNotExist 	"data/to0/h/i/j/something"
assertExist 	"data/to1/h/i/j/something"


# make a dir in the deep dir of another server
echo 
echo "make a dir in the deep dir of another server"


# delete a dir int he deepdir of another server
echo 
echo "delete a dir int he deepdir of another server"

stop_jobs
pass_test   
