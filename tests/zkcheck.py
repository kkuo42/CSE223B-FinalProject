import json
from subprocess import check_output
import sys
import os

# this is a dict of all of the different json responses we may get
file_check = {"a_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "Addr": "localhost:9500"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 0, "ReadCount": 0, "Addr": "localhost:9601"}}},
	      "b_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "Addr": "localhost:9501"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 0, "Addr": "localhost:9602"}}},
	      "c_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "Addr": "localhost:9502"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 0, "Addr": "localhost:9600"}}},
	      "a_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 6, "Addr": "localhost:9500"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 0, "ReadCount": 3, "Addr": "localhost:9601"},
				           "localhost:9602": {"WriteCount": 0, "ReadCount": 1, "Addr": "localhost:9602"}}},
	      "b_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 11, "Addr": "localhost:9501"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 5, "Addr": "localhost:9602"},
				           "localhost:9600": {"WriteCount": 0, "ReadCount": 3, "Addr": "localhost:9600"}}},
	      "c_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 11, "Addr": "localhost:9502"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 50, "Addr": "localhost:9600"},
				           "localhost:9601": {"WriteCount": 0, "ReadCount": 3, "Addr": "localhost:9601"}}},
	      "a_data_fw": {"Primary": {"WriteCount": 6, "ReadCount": 6, "Addr": "localhost:9500"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 10, "ReadCount": 3, "Addr": "localhost:9601"},
				           "localhost:9602": {"WriteCount": 2, "ReadCount": 1, "Addr": "localhost:9602"}}},
	      "b_data_fw": {"Primary": {"WriteCount": 11, "ReadCount": 11, "Addr": "localhost:9501"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 5, "ReadCount": 5, "Addr": "localhost:9602"},
				           "localhost:9600": {"WriteCount": 3, "ReadCount": 3, "Addr": "localhost:9600"}}},
	      "c_data_fw": {"Primary": {"WriteCount": 8, "ReadCount": 11, "Addr": "localhost:9502"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 10, "ReadCount": 50, "Addr": "localhost:9600"},
				           "localhost:9601": {"WriteCount": 1, "ReadCount": 3, "Addr": "localhost:9601"}}}
	     }

server_check = {
		
	       }

op = sys.argv[1]
path = sys.argv[2]
check = sys.argv[3]

FNULL = open(os.devnull, 'w')
out = check_output(['zkcli',op,path,check], stderr=FNULL)
data = json.loads(out)

# handle different types of data
if "data" in path:
	# we will be checking the json 
	if data["Primary"] == file_check[check]["Primary"] and data["Replicas"] == file_check[check]["Replicas"]:
		print "pass"
	else:
		print "zkdata:", data["Primary"], data["Replicas"]
		print "checkdata:", file_check[check]
		print "fail"
elif "alivemeta" in path:
	pass
