# #!/bin/bash

# Before using this script, all servers must be setup with the aws script.
# In addition, ssh keys must be generated to login without a prompt.

# Instructions:
# "ssh-keygen" and then press enter at every prompt  
# "ssh-copy-id <remote hostname>" 

# This script will spawn zookeper on the first server specified.
# It will then run fs-server on each of the specified servers.

if [ "$#" -ne 1 ]; then
    echo "Usage: setupremote.sh <server#>"
	exit 1
fi

# build source
make || exit 1
echo

server1=ubuntu@ec2-34-219-144-16.us-west-2.compute.amazonaws.com
server2=ubuntu@ec2-13-229-209-89.ap-southeast-1.compute.amazonaws.com
server3=ubuntu@ec2-13-125-98-158.ap-northeast-2.compute.amazonaws.com
#server4=ubuntu@ec2-52-18-39-217.eu-west-1.compute.amazonaws.com
#server5=ubuntu@ec2-52-67-217-162.sa-east-1.compute.amazonaws.com

key1=keys/cse223b-project-group2-us-west-2.pem 
key2=keys/cse223b-project-group2-ap-southeast-1.pem 
key3=keys/cse223b-project-group2-ap-northeast-2.pem 
#key4=keys/cse223b-project-group2-eu-west-1.pem 
#key5=keys/cse223b-project-group2-sa-east-1.pem 


# teardown
stop_jobs() {
echo "killing anything started"
for ((i=1; i<=3; i++)); do
servernum=server$i
keynum=key$i
hostname=${!servernum}
zkkeyl=${!keynum}
ssh -i $zkkeyl $hostname <<EOSHH
echo "pkill"
pkill fs-server
pkill fs-front
fusermount -u go/src/proj/data/to$servernum
EOSHH
done
}
		

servernum=server$1
keynum=key$1
zkserver=${!servernum}
zkkey=${!keynum}

ssh -i $zkkey $zkserver <<EOSHH
pkill java
cd go/src/proj
rm -rf zkdata zookeeper.out
mkdir zkdata
zookeeper-3.4.12/bin/zkServer.sh start
rm -rf data logs
mkdir data logs
EOSHH

for ((i=1; i<=3; i++)); do
servernum=server$i
keynum=key$i
hostname=${!servernum}
zkkeyl=${!keynum}
ssh -i $zkkeyl $hostname <<EOSHH
pkill fs-server
pkill fs-front
cd go/src/proj
rm data/to$servernum/*
rm data/from$servernum/*
fusermount -u data/to$servernum
make
mkdir data logs
mkdir -p data/from$servernum
mkdir -p data/to$servernum
rm -f logs/$servernum.txt
rm -f logs/front$servernum.txt
fs-server data/from$servernum >> logs/$servernum.txt & serverPID=\$!
sleep 1
echo
if !( ps -p \$serverPID > /dev/null ) then 
echo
echo 'Failed to lauch a server'
echo
fi
sleep 5
fs-front data/to$servernum >> logs/front$servernum.txt & clientPID=\$!
sleep 1
echo
if !( ps -p \$clientPID > /dev/null ) then 
echo
echo 'Failed to lauch a client'
echo
exit 1
fi
exit 0
EOSHH
done

echo

#wait for shutdown
while true; do
    read -p "Do you wish to stop all servers and clients this program? [y/n] " yn
    case $yn in
        [Yy]* ) 
		stop_jobs
		ssh -i $zkkey $zkserver "go/src/proj/zookeeper-3.4.12/bin/zkServer.sh stop"
        exit 0
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
