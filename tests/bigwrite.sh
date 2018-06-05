#!/bin/bash

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo -n "init" > data/to0/a
echo -n "init" > acheck

#cat data/to0/a
#cat data/to1/a
cat data/to2/a

echo "doing 5 writes on a from serv 0, 10 from serv 1, 2 from serv 2"
for ((n=0;n<5;n++))
do 
	echo "serv 0"
	echo "serv 0" >> data/to0/a
	echo "serv 0" >> acheck
done
for ((n=0;n<10;n++))
do 
	echo "serv 1" >> data/to1/a
	echo "serv 1" >> acheck
done
for ((n=0;n<2;n++))
do 
	echo "serv 2" >> data/to2/a
	echo "serv 2" >> acheck
done
assertFilesEqual data/from0/a acheck
rm acheck

stop_jobs
pass_test
