import os
import time
import multiprocessing

WRITE_NUM = 25
def writefile(value, path, return_dict):
	start_time = time.time()
	for i in range(WRITE_NUM):
		os.system("echo -n "+value+" >> "+path)
	elapsed_time = time.time() - start_time
	return_dict[path] = elapsed_time

manager = multiprocessing.Manager()
return_dict = manager.dict()	

os.system("rm data/to0/a")
os.system("echo -n a > data/to0/a")

p1 = multiprocessing.Process(target=writefile, args=("a", "data/to0/a", return_dict))
p2 = multiprocessing.Process(target=writefile, args=("a", "data/to1/a", return_dict))
p3 = multiprocessing.Process(target=writefile, args=("a", "data/to2/a", return_dict))
p4 = multiprocessing.Process(target=writefile, args=("a", "data/to3/a", return_dict))
pArray = [p1,p2,p3,p4]
for p in pArray:
	p.start()
for p in pArray:
	p.join()
print WRITE_NUM,"writes to the same file from, 4 different clients:"
for k,v in return_dict.items():
	print "\tProcess:", k, "total time:", v
	print "\twrites/second:\t", WRITE_NUM/v

os.system("rm data/to0/a")

return_dict = manager.dict()	

start_time = time.time()
p1 = multiprocessing.Process(target=writefile, args=("a", "data/to0/a1", return_dict))
p2 = multiprocessing.Process(target=writefile, args=("b", "data/to1/b1", return_dict))
p3 = multiprocessing.Process(target=writefile, args=("c", "data/to2/c1", return_dict))
p4 = multiprocessing.Process(target=writefile, args=("d", "data/to3/d1", return_dict))
pArray = [p1,p2,p3,p4]
for p in pArray:
	p.start()
for p in pArray:
	p.join()
elapsed_time = time.time() - start_time
print WRITE_NUM*4, "parallel write, 4 different clients:"
for k,v in return_dict.items():
	print "\tProcess:", k, "total time:", v
	print "\twrites/second:\t", WRITE_NUM/v
os.system("rm data/to0/a1")
os.system("rm data/to1/b1")
os.system("rm data/to2/c1")
os.system("rm data/to3/d1")

start_time = time.time()
for i in range(WRITE_NUM):
	os.system("echo -n a >> data/to0/a2")
for i in range(WRITE_NUM*3):
	os.system("echo -n b >> data/to1/b2")
elapsed_time = time.time() - start_time
print WRITE_NUM, "clientto0 writes then", WRITE_NUM*3, "clientto1 writes:"
print "\ttotal time:\t", elapsed_time, "seconds"
print "\twrites/second:\t", WRITE_NUM*4/elapsed_time
os.system("rm data/to0/a2")
os.system("rm data/to1/b2")

