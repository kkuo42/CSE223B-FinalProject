# #!/bin/bash

# build source
make || exit 1
echo

server0=cse223b_kjkuo@vm166.sysnet.ucsd.edu
server1=cse223b_kjkuo@vm167.sysnet.ucsd.edu
server2=cse223b_kjkuo@vm168.sysnet.ucsd.edu
server3=cse223b_kjkuo@vm169.sysnet.ucsd.edu
server4=cse223b_kjkuo@vm170.sysnet.ucsd.edu
server5=cse223b_kjkuo@vm171.sysnet.ucsd.edu
server6=cse223b_kjkuo@vm172.sysnet.ucsd.edu
server7=cse223b_kjkuo@vm173.sysnet.ucsd.edu

ssh $server0 < tests/startzk.sh

#for hostname in $server0 $server1 $server2 $server3 $server4 $server5 $server6 $server7; do
	ssh $hostname < tests/startserver.sh
#done

ssh $server0 < tests/stopzk.sh
