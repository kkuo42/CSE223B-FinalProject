# #!/bin/bash

# Before using this script, all servers must be setup with the aws script.
# In addition, ssh keys must be generated to login without a prompt.

# Instructions:
# "ssh-keygen" and then press enter at every prompt  
# "ssh-copy-id <remote hostname>" 

# This script will spawn zookeper on the first server specified.
# It will then run fs-server on each of the specified servers.

server0=cse223b_kjkuo@vm166.sysnet.ucsd.edu
server1=cse223b_kjkuo@vm167.sysnet.ucsd.edu
server2=cse223b_kjkuo@vm168.sysnet.ucsd.edu
server3=cse223b_kjkuo@vm169.sysnet.ucsd.edu
server4=cse223b_kjkuo@vm170.sysnet.ucsd.edu
server5=cse223b_kjkuo@vm171.sysnet.ucsd.edu
server6=cse223b_kjkuo@vm172.sysnet.ucsd.edu
server7=cse223b_kjkuo@vm173.sysnet.ucsd.edu

ssh -q $server0 <<EOSHH
cd gopath/src/proj
rm -rf zkdata zookeeper.out
mkdir zkdata
zookeeper-3.4.12/bin/zkServer.sh start
EOSHH

for hostname in $server0 $server1 $server2 $server3 $server4 $server5 $server6 $server7; do
ssh -q $hostname <<EOSHH
cd gopath/src/proj
make
rm -rf from$hostname
mkdir from$hostname
rm $hostname.txt
fs-server from$hostname >> $hostname.txt & serverPID=\$!
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

#wait for shutdown
while true; do
    read -p "Do you wish to stop all servers and clients this program? [y/n] " yn
    case $yn in
        [Yy]* ) 
		ssh -q $server0 "gopath/src/proj/zookeeper-3.4.12/bin/zkServer.sh stop"
        exit 0
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
