# #!/bin/bash

# Before using this script, all servers must be setup with the aws script.
# In addition, ssh keys must be generated to login without a prompt.

# Instructions:
# "ssh-keygen" and then press enter at every prompt  
# "ssh-copy-id <remote hostname>" 

# This script will spawn zookeper on the first server specified.
# It will then run fs-server on each of the specified servers.

if [ "$#" -ne 2 ]; then
    echo "Usage: setupremote.sh <username> <server#>"
	exit 1
fi

# teardown func
stop_jobs() {
	echo "killing anything started"
	kill $front0PID
	kill $front1PID
	kill $front2PID
	kill $front3PID
    #OSX Unmount
    unamestr=`uname`
    if [[ "$unamestr" == 'Darwin' ]]; then
        echo "OSX UNMOUNT"
        umount data/to0
        umount data/to1
        umount data/to2
        umount data/to3
    else 
        fusermount -u data/to0
        fusermount -u data/to1
        fusermount -u data/to2
        fusermount -u data/to3
    fi
	echo "killed all processes"
}

# build source
make || exit 1
echo

username=$1

server0=cse223b_${username}@vm166.sysnet.ucsd.edu
server1=cse223b_${username}@vm167.sysnet.ucsd.edu
server2=cse223b_${username}@vm168.sysnet.ucsd.edu
server3=cse223b_${username}@vm169.sysnet.ucsd.edu
server4=cse223b_${username}@vm170.sysnet.ucsd.edu
server5=cse223b_${username}@vm171.sysnet.ucsd.edu
server6=cse223b_${username}@vm172.sysnet.ucsd.edu
server7=cse223b_${username}@vm173.sysnet.ucsd.edu

servernum=server$2
zkserver=${!servernum}

ssh -q $zkserver <<EOSHH
pkill java
cd gopath/src/proj
rm -rf zkdata zookeeper.out
mkdir zkdata
zookeeper-3.4.12/bin/zkServer.sh start
rm -rf data logs
mkdir data logs
EOSHH

for hostname in $server0 $server1 $server2 $server3 $server4 $server5 $server6 $server7; do
ssh -q $hostname <<EOSHH
cd gopath/src/proj
mkdir data/from$hostname
rm -f logs/$hostname.txt
fs-server data/from$hostname >> logs/$hostname.txt & serverPID=\$!
sleep 1
echo
if !( ps -p \$serverPID > /dev/null ) then 
echo
echo 'Failed to lauch a server'
echo
exit 1
fi
exit 0
EOSHH
done

echo

# reset data and logs
rm -rf data
mkdir data
rm -rf logs
mkdir logs

mkdir data/to0

fs-front data/to0 >> logs/front0_log.txt &
front0PID=$!
sleep 1
echo

if !( ps -p $front0PID > /dev/null )
then
   	stop_jobs
	echo
   	echo "Failed to lauch a client 0"
   	echo
   	exit 1
fi

mkdir data/to1

fs-front data/to1 >> logs/front1_log.txt &
front1PID=$!
sleep 1
echo

if !( ps -p $front1PID > /dev/null )
then
   	stop_jobs
	echo
   	echo "Failed to lauch a client 1"
   	echo
   	exit 1
fi

mkdir data/to2

fs-front data/to2 >> logs/front2_log.txt &
front2PID=$!
sleep 1
echo

if !( ps -p $front2PID > /dev/null )
then
   	stop_jobs
	echo
   	echo "Failed to lauch a client 2"
   	echo
   	exit 1
fi

mkdir data/to3

fs-front data/to3 >> logs/front3_log.txt &
front3PID=$!
sleep 1
echo

if !( ps -p $front3PID > /dev/null )
then
   	stop_jobs
	echo
   	echo "Failed to lauch a client 3"
   	echo
   	exit 1
fi

#wait for shutdown
while true; do
    read -p "Do you wish to stop all servers and clients this program? [y/n] " yn
    case $yn in
        [Yy]* ) 
   		stop_jobs
		ssh -q $zkserver "gopath/src/proj/zookeeper-3.4.12/bin/zkServer.sh stop"
        exit 0
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
