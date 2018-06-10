import os
import time
import multiprocessing

WRITE_NUM = 5

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a2")
for i in range(WRITE_NUM*3):
	os.system("echo -n b >> data/to1/a2")
elapsed_time = time.time() - start_time
print WRITE_NUM, "clientto0 writes then", WRITE_NUM*3, "clientto1 writes:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time

time.sleep(60)

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a2")
for i in range(WRITE_NUM*3):
	os.system("echo -n b >> data/to1/a2")
elapsed_time = time.time() - start_time
print WRITE_NUM, "clientto0 writes then", WRITE_NUM*3, "clientto1 writes:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time
