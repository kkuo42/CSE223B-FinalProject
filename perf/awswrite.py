import os
import time

WRITE_NUM = 10

start_time = time.time()
for i in range(WRITE_NUM):
	print ("writing ", i, "th time")
	os.system("echo -n a >> data/toserver1/a")
elapsed_time = time.time() - start_time
print ("serial write:")
print ("\ttotal time:\t", elapsed_time, "seconds")
print ("\twrites/second:\t", WRITE_NUM/elapsed_time)

