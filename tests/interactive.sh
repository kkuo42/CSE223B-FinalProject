# #!/bin/bash

# this script creates a keeper locally. 
# it expects zookeeper to be in the same directory
# with the datadir in zookeeper-3.4.12/conf/zoo.cfg set to zkdata also in this directory

# the script lauches clients+servers as pairs so that the client is guaranteed to be assigned 
# to that server because the frontend will assign most recently joined server

source tests/functions.sh

setup_servers

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

