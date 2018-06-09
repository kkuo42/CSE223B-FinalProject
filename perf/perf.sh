# #!/bin/bash

# http://www-01.ibm.com/support/docview.wss?uid=swg21577115
# sudo apt install iozone3
# or
# download and make http://www.iozone.org/src/current/iozone3_482.tar

if [ "$#" -ne 1 ]; then
    echo "Usage: perf.sh <output_prefix> "
	exit 1
fi

output_prefix=$1


# single
iozone -R -l 1 -u 1 -r 4k -s 512 -i 0 -i 1 -b ${output_prefix}_1.wks -F ../data/to0/temp

# 2
iozone -R -l 2 -u 2 -r 4k -s 512 -i 0 -i 1 -b ${output_prefix}_2.wks -F ../data/to0/temp ../data/to1/temp

# 4
iozone -R -l 4 -u 4 -r 4k -s 512 -i 0 -i 1 -b ${output_prefix}_4.wks -F ../data/to0/temp ../data/to1/temp ../data/to2/temp ../data/to3/temp
