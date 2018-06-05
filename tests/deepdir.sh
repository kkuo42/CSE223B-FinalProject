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
sleep 1

assertExist 	"data/from0/a/b/c" # check that nesting works
assertExist 	"data/to1/a/b/c"
assertExist 	"data/to2/a/b/c"

echo "writing something to data/to0/a/b/c/something"
echo "something" >> data/to0/a/b/c/something
sleep 1

echo "cat data/to0/a/b/c/something"
cat data/to0/a/b/c/something > /dev/null
sleep 1

assertExist 	"data/from0/a/b/c/something"
assertExist 	"data/from1/a/b/c/something"
assertNotExist 	"data/from2/a/b/c/something"
assertExist 	"data/to0/a/b/c/something"
assertExist 	"data/to1/a/b/c/something"
assertExist 	"data/to2/a/b/c/something"

echo "cat data/to2/a/b/c/something"
cat data/to2/a/b/c/something > /dev/null
sleep 1

assertExist "data/from0/a/b/c/something"
assertExist "data/from1/a/b/c/something"
assertExist "data/from2/a/b/c/something"
assertExist "data/to0/a/b/c/something"
assertExist "data/to1/a/b/c/something"
assertExist "data/to2/a/b/c/something"

assertFilesEqual "data/to0/a/b/c/something" "data/to1/a/b/c/something"


# write a file in the deep dir of another server
echo
echo "write a file in the deep dir of another server"

echo "creating nested directories data/to0/e/f/g"
mkdir -p data/to0/e/f/g
sleep 1

echo "writing something to data/to0/e/f/g/something"
echo "something" >> data/to0/e/f/g/something
sleep 1

echo "writing something else to data/to1/e/f/g/something"
echo "something else" >> data/to1/e/f/g/something
sleep 1

assertExist "data/from0/e/f/g/something"
assertExist "data/from1/e/f/g/something"
assertNotExist "data/from2/e/f/g/something"
assertExist "data/to0/e/f/g/something"
assertExist "data/to1/e/f/g/something"
assertExist "data/to2/e/f/g/something"

touch temp
echo "something" >> temp
echo "something else" >> temp
assertFilesEqual "data/to0/e/f/g/something" "temp"
assertFilesEqual "data/to1/e/f/g/something" "temp"
rm temp


# create a file in the deep dir of another server
echo 
echo "create a file in the deep dir of another server"

echo "creating nested directories data/to0/h/i/j"
mkdir -p data/to0/h/i/j
sleep 1

echo "writing something else to data/to1/h/i/j/something"
echo "something" >> data/to1/h/i/j/something
sleep 1

assertNotExist 	"data/from0/h/i/j/something"
assertExist 	"data/from1/h/i/j/something"
assertExist 	"data/from2/h/i/j/something"
assertExist 	"data/to0/h/i/j/something"
assertExist 	"data/to1/h/i/j/something"
assertExist 	"data/to2/h/i/j/something"


# rename a file int he deepdir of another server
echo 
echo "rename a file int he deepdir of another server"

echo "creating nested directories data/to0/v/w/x"
mkdir -p data/to0/v/w/x
sleep 1

echo "writing something to data/to0/v/w/x/something"
echo "something" >> data/to0/v/w/x/something
sleep 1

# replica rename
echo "moving something to data/to1/v/w/x/else"
mv data/to1/v/w/x/something data/to1/v/w/x/else
sleep 1

assertExist 	"data/from0/v/w/x/else"
assertExist 	"data/from1/v/w/x/else"
assertNotExist 	"data/from2/v/w/x/else"
assertExist 	"data/to0/v/w/x/else"
assertExist 	"data/to1/v/w/x/else"
assertExist 	"data/to2/v/w/x/else"


# make a dir in the deep dir of another server
echo 
echo "make a dir in the deep dir of another server"

echo "creating nested directories data/to0/l/m/n"
mkdir -p data/to0/l/m/n
sleep 1

echo "creating directory data/to1/l/m/n/o"
mkdir data/to1/l/m/n/o
sleep 1

assertNotExist 	"data/from0/l/m/n/o"
assertExist 	"data/from1/l/m/n/o"
assertExist 	"data/from2/l/m/n/o"
assertExist 	"data/to0/l/m/n/o"
assertExist 	"data/to1/l/m/n/o"
assertExist 	"data/to2/l/m/n/o"

# delete a dir int he deepdir of another server
echo 
echo "delete a dir int he deepdir of another server"

echo "creating nested directories data/to0/p/q/r"
mkdir -p data/to0/p/q/r
sleep 1

echo "removing directory data/to1/p/q/r"
rmdir data/to1/p/q/r
sleep 1

assertNotExist 	"data/from0/p/q/r"
assertNotExist 	"data/from1/p/q/r"
assertNotExist 	"data/from2/p/q/r"
assertNotExist 	"data/to0/p/q/r"
assertNotExist 	"data/to1/p/q/r"
assertNotExist 	"data/to2/p/q/r"


# rename a dir int he deepdir of another server
echo 
echo "delete a dir int he deepdir of another server"

echo "creating nested directories data/to0/s/t/u"
mkdir -p data/to0/s/t/u
sleep 1

echo "moving directory u to data/to1/s/t/v"
mv data/to1/s/t/u data/to1/s/t/v
sleep 1

assertExist 	"data/from0/s/t/v"
assertExist 	"data/from1/s/t/v"
assertNotExist 	"data/from2/s/t/v"
assertExist 	"data/to0/s/t/v"
assertExist 	"data/to1/s/t/v"
assertExist 	"data/to2/s/t/v"

stop_jobs
pass_test   
