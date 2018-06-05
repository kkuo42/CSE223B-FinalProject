# #!/bin/bash

source tests/functions.sh
testName=$(basename $BASH_SOURCE)

setup_servers

echo "reading from nonexistent file a"
cat data/to0/a
sleep 1
echo "stuff"

stop_jobs
pass_test
