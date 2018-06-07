cd gopath/src/proj
make
rm -rf from$hostname
mkdir from$hostname
fs-server from >> $hostname.txt & server0PID=$!
sleep 1
echo
