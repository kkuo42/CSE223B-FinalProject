import os
import time

WRITE_NUM = 50

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a")
	os.system("echo -n b >> data/to1/a")
	os.system("echo -n c >> data/to2/a")
	os.system("echo -n d >> data/to3/a")
elapsed_time = time.time() - start_time
print WRITE_NUM*4,"serial write:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time

os.system("rm data/to0/a")

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a1")
	os.system("echo -n b >> data/to1/b1")
	os.system("echo -n c >> data/to2/c1")
	os.system("echo -n d >> data/to3/d1")
elapsed_time = time.time() - start_time
print WRITE_NUM*4, "parallel write:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a2")
for i in range(WRITE_NUM*3):
	os.system("echo -n b >> data/to1/b2")
elapsed_time = time.time() - start_time
print WRITE_NUM, "clientto0", WRITE_NUM*3, "clientto1 write:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time

