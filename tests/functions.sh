# #!/bin/bash

# This script is used to define all of the functions that will be used in other tests

setup_servers() {
	# build source
	make || exit 1
	echo

	# zk reset
	rm -rf zkdata zookeeper.out
	mkdir zkdata
	zookeeper-3.4.12/bin/zkServer.sh start
	echo

	# reset data and logs
	rm -rf data
	mkdir data
	rm -rf logs
	mkdir logs

	# pair 0
	mkdir data/from0 data/to0

	fs-server data/from0 localhost:9500 >> logs/server0_log.txt &
	server0PID=$!
	sleep 1
	echo

	fs-front data/to0 localhost:9500 >> logs/front0_log.txt &
	front0PID=$!
	sleep 1
	echo

	if !( ( ps -p $server0PID > /dev/null ) && ( ps -p $front0PID > /dev/null ) )
	then
    	stop_jobs
		echo
    	echo "Failed to lauch a client/server pair 0"
    	echo
    	exit 1
	fi

	# pair 1
	mkdir data/from1 data/to1

	fs-server data/from1 localhost:9501 >> logs/server1_log.txt&
	server1PID=$!
	sleep 1
	echo

	fs-front data/to1 localhost:9501 >> logs/front1_log.txt&
	front1PID=$!
	sleep 1
	echo

	if !( ( ps -p $server1PID > /dev/null ) && ( ps -p $front1PID > /dev/null ) )
	then
	    stop_jobs
		echo
	    echo "Failed to lauch a client/server pair 1"
	    echo
	    exit 1
	fi
}

# teardown func
stop_jobs() {
	echo "killing anything started"
	zookeeper-3.4.12/bin/zkServer.sh stop
	kill $server0PID
	kill $server1PID
	kill $front0PID
	kill $front1PID
    #OSX Unmount
    unamestr=`uname`
    if [[ "$unamestr" == 'Darwin' ]]; then
        echo "OSX UNMOUNT"
        umount data/to0
        umount data/to1
    else 
        fusermount -u data/to0
        fusermount -u data/to1
    fi
	echo "killed all processes"
}

# let user name test passed
pass_test() {
	echo
	echo "PASSED $testName"
	echo
}

# assert file contents, 1st arg filename, 2nd arg expected contents
assertFile() {
	echo -n $2 > data/temp
	if cmp -s data/temp $1
	then
		rm data/temp
		echo "PASSED assertFile. $1 contents: $2"
	else
		rm data/temp
		echo "FAILED assertFile. $1 contents: $2"
		echo "Exiting."
		echo
		stop_jobs
		tail logs/*.txt
		echo
		echo "FAILED $testName"
		echo
		exit 1
	fi
}

# assert file existance, 1st arg filename
assertExist() {
	if [ -e $1 ]
	then
		echo "PASSED assertExist. $1"
	else
		echo "FAILED assertExist. $1"
		echo "Exiting."
		echo
		stop_jobs
		tail logs/*.txt
		echo
		echo "FAILED $testName"
		echo
		exit 1
	fi
}
# assert no file existance, 1st arg filename
assertNotExist() {
	if [ -e $1 ]
	then
		echo "FAILED assertNotExist. $1"
		echo "Exiting."
		echo
		stop_jobs
		tail logs/*.txt
		echo
		echo "FAILED $testName"
		echo
		exit 1
	else
		echo "PASSED assertNotExist. $1"
	fi
}

