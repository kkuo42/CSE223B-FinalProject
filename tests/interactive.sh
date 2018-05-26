# #!/bin/bash

# this script creates a keeper locally. 
# it expects zookeeper to be in the same directory
# with the datadir in zookeeper-3.4.12/conf/zoo.cfg set to zkdata also in this directory

# the script lauches clients+servers as pairs so that the client is guaranteed to be assigned 
# to that server because the frontend will assign most recently joined server


# build source
make || exit 1
echo

# teardown func
stop_jobs() {
	echo "killing anything started"
	zookeeper-3.4.12/bin/zkServer.sh stop
	echo
	kill $server0PID
	kill $server1PID
	echo
	kill $front0PID
	kill $front1PID
	echo
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
	echo
	
	tail logs/*.txt
}


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

if !( ( ps -p $server0PID > /dev/null ) && ( ps -p $server0PID > /dev/null ) )
then
	echo
    echo "Failed to lauch a client/server pair 0"
    echo
    stop_jobs
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

if !( ( ps -p $server0PID > /dev/null ) && ( ps -p $server0PID > /dev/null ) )
then
	echo
    echo "Failed to lauch a client/server pair 1"
    echo
    stop_jobs
    exit 1
fi

#wait for shutdown
while true; do
    read -p "Do you wish to stop all servers and clients this program? [y/n] " yn
    case $yn in
        [Yy]* ) 
        stop_jobs
        exit 0
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done

