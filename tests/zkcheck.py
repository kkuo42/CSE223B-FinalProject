import json
from subprocess import check_output
import sys
import os

# this is a dict of all of the different json responses we may get
file_check = {"a_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}}},
	      "b_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9501",  "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}}},
	      "c_data_init": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9502",  "SFSAddr": "localhost:9602"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9500",  "SFSAddr": "localhost:9600"}}},
	      "a_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 6, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 0, "ReadCount": 3, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"},
				           "localhost:9602": {"WriteCount": 0, "ReadCount": 1, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}}},
	      "b_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 11, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 5, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"},
				           "localhost:9600": {"WriteCount": 0, "ReadCount": 3, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}}},
	      "c_data_fr": {"Primary": {"WriteCount": 1, "ReadCount": 11, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 50, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"},
				           "localhost:9601": {"WriteCount": 0, "ReadCount": 3, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}}},
	      "a_data_fw": {"Primary": {"WriteCount": 6, "ReadCount": 6, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 10, "ReadCount": 3, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"},
				           "localhost:9602": {"WriteCount": 2, "ReadCount": 1, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}}},
	      "b_data_fw": {"Primary": {"WriteCount": 11, "ReadCount": 11, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 5, "ReadCount": 5, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"},
				           "localhost:9600": {"WriteCount": 3, "ReadCount": 3, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}}},
	      "c_data_fw": {"Primary": {"WriteCount": 8, "ReadCount": 11, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 10, "ReadCount": 50, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"},
				           "localhost:9601": {"WriteCount": 1, "ReadCount": 3, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}}},
	      "a_data_fb": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}}},
	      "b_data_fb": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}}},
	      "c_data_fb": {"Primary": {"WriteCount": 2, "ReadCount": 0, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9600": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9500", "SFSAddr": "localhost:9600"}}},
	      "a_data_fp": {"Primary": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}}},
	      "b_data_fp": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}, 
			      "Replicas": {"localhost:9602": {"WriteCount": 0, "ReadCount": 0, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}}},
	      "c_data_fp": {"Primary": {"WriteCount": 1, "ReadCount": 1, "CoordAddr": "localhost:9502", "SFSAddr": "localhost:9602"}, 
			      "Replicas": {"localhost:9601": {"WriteCount": 2, "ReadCount": 0, "CoordAddr": "localhost:9501", "SFSAddr": "localhost:9601"}}}
	     }

server_check = {
		"9500_init": {"PrimaryFor": {"a":"a"}, "ReplicaFor": {}},
		"9501_init": {"PrimaryFor": {"b":"b"}, "ReplicaFor": {}},
		"9502_init": {"PrimaryFor": {"c":"c"}, "ReplicaFor": {}},
		"9600_init": {"PrimaryFor": {}, "ReplicaFor": {"c":"c"}},
		"9601_init": {"PrimaryFor": {}, "ReplicaFor": {"a":"a"}},
		"9602_init": {"PrimaryFor": {}, "ReplicaFor": {"b":"b"}},
		"9500_r": {"PrimaryFor": {"a":"a"}, "ReplicaFor": {}},
		"9501_r": {"PrimaryFor": {"b":"b"}, "ReplicaFor": {}},
		"9502_r": {"PrimaryFor": {"c":"c"}, "ReplicaFor": {}},
		"9600_r": {"PrimaryFor": {}, "ReplicaFor": {"c":"c", "b":"b"}},
		"9601_r": {"PrimaryFor": {}, "ReplicaFor": {"a":"a", "c":"c"}},
		"9602_r": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "a":"a"}},
		"9500_mv": {"PrimaryFor": {"anew":"anew"}, "ReplicaFor": {}},
		"9501_mv": {"PrimaryFor": {"b":"b"}, "ReplicaFor": {}},
		"9502_mv": {"PrimaryFor": {"c":"c"}, "ReplicaFor": {}},
		"9600_mv": {"PrimaryFor": {}, "ReplicaFor": {"c":"c", "b":"b"}},
		"9601_mv": {"PrimaryFor": {}, "ReplicaFor": {"anew":"anew", "c":"c"}},
		"9602_mv": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "anew":"anew"}},
		"9500_mv2": {"PrimaryFor": {"anew":"anew"}, "ReplicaFor": {}},
		"9501_mv2": {"PrimaryFor": {"bnew":"bnew"}, "ReplicaFor": {}},
		"9502_mv2": {"PrimaryFor": {"c":"c"}, "ReplicaFor": {}},
		"9600_mv2": {"PrimaryFor": {}, "ReplicaFor": {"c":"c", "bnew":"bnew"}},
		"9601_mv2": {"PrimaryFor": {}, "ReplicaFor": {"anew":"anew", "c":"c"}},
		"9602_mv2": {"PrimaryFor": {}, "ReplicaFor": {"bnew":"bnew", "anew":"anew"}},
		"9500_mv3": {"PrimaryFor": {"anew":"anew"}, "ReplicaFor": {}},
		"9501_mv3": {"PrimaryFor": {"bnew":"bnew"}, "ReplicaFor": {}},
		"9502_mv3": {"PrimaryFor": {"cnew":"cnew"}, "ReplicaFor": {}},
		"9600_mv3": {"PrimaryFor": {}, "ReplicaFor": {"cnew":"cnew", "bnew":"bnew"}},
		"9601_mv3": {"PrimaryFor": {}, "ReplicaFor": {"anew":"anew", "cnew":"cnew"}},
		"9602_mv3": {"PrimaryFor": {}, "ReplicaFor": {"bnew":"bnew", "anew":"anew"}},
		"9500_rm1": {"PrimaryFor": {"anew":"anew"}, "ReplicaFor": {}},
		"9501_rm1": {"PrimaryFor": {"bnew":"bnew"}, "ReplicaFor": {}},
		"9502_rm1": {"PrimaryFor": {}, "ReplicaFor": {}},
		"9600_rm1": {"PrimaryFor": {}, "ReplicaFor": {"bnew":"bnew"}},
		"9601_rm1": {"PrimaryFor": {}, "ReplicaFor": {"anew":"anew"}},
		"9602_rm1": {"PrimaryFor": {}, "ReplicaFor": {"bnew":"bnew", "anew":"anew"}},
		"9500_f1": {"PrimaryFor": {"a":"a"}, "ReplicaFor": {}},
		"9501_f1": {"PrimaryFor": {"b":"b", "c":"c"}, "ReplicaFor": {}},
		"9600_f1": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "c":"c"}},
		"9601_f1": {"PrimaryFor": {}, "ReplicaFor": {"a":"a"}},
		"9501_f2": {"PrimaryFor": {"b":"b", "a":"a"}, "ReplicaFor": {}},
		"9502_f2": {"PrimaryFor": {"c":"c"}, "ReplicaFor": {}},
		"9601_f2": {"PrimaryFor": {}, "ReplicaFor": {"c":"c"}},
		"9602_f2": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "a":"a"}},
		"9501_f3": {"PrimaryFor": {"b":"b", "a":"a", "c":"c", "c2":"c2"}, "ReplicaFor": {}},
		"9502_f3": {"PrimaryFor": {"c3":"c3"}, "ReplicaFor": {}},
		"9601_f3": {"PrimaryFor": {}, "ReplicaFor": {"c3":"c3"}},
		"9602_f3": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "a":"a", "c":"c", "c2":"c2"}},
		"9501_f32": {"PrimaryFor": {"b":"b", "a":"a", "c":"c"}, "ReplicaFor": {}},
		"9502_f32": {"PrimaryFor": {"c3":"c3", "c2":"c2"}, "ReplicaFor": {}},
		"9601_f32": {"PrimaryFor": {}, "ReplicaFor": {"c3":"c3", "c2":"c2"}},
		"9602_f32": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "a":"a", "c":"c"}},
		"9500_ff1": {"PrimaryFor": {"a":"a", "c2":"c2"}, "ReplicaFor": {}},
		"9501_ff1": {"PrimaryFor": {"b":"b", "c":"c"}, "ReplicaFor": {}},
		"9600_ff1": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "c":"c"}},
		"9601_ff1": {"PrimaryFor": {}, "ReplicaFor": {"a":"a", "c2":"c2"}},
		"9500_ff2": {"PrimaryFor": {"a":"a"}, "ReplicaFor": {}},
		"9501_ff2": {"PrimaryFor": {"b":"b", "c":"c", "c2":"c2"}, "ReplicaFor": {}},
		"9600_ff2": {"PrimaryFor": {}, "ReplicaFor": {"b":"b", "c":"c", "c2":"c2"}},
		"9601_ff2": {"PrimaryFor": {}, "ReplicaFor": {"a":"a"}}
	       }

alive_check = {
		"alive_b1": ["localhost:9500", "localhost:9501", "localhost:9502", "localhost:9600","localhost:9601","localhost:9602"],
		"alive_a1": ["localhost:9500", "localhost:9501", "localhost:9600", "localhost:9601"],
		"alive_a2": ["localhost:9501", "localhost:9502", "localhost:9601", "localhost:9602"]
		}

op = sys.argv[1]
path = sys.argv[2]
check = sys.argv[3]
c2 = False

if len(sys.argv) == 5:
	c2 = True
	check2 = sys.argv[4]

FNULL = open(os.devnull, 'w')
out = check_output(['zkcli',op,path,check], stderr=FNULL)
if op != "ls":
	data = json.loads(out)

# handle different types of data
if "data" in path and "get" in op:
	# we will be checking the json 
	if data["Primary"] == file_check[check]["Primary"] and data["Replicas"] == file_check[check]["Replicas"]:
		print "pass"
	elif c2 and data["Primary"] == file_check[check2]["Primary"] and data["Replicas"] == file_check[check2]["Replicas"]:
		print "pass"
	else:
		print "zkdata:", data["Primary"], data["Replicas"]
		print "checkdata:", file_check[check]
		if c2:
			print "checkdata2:", file_check[check2]
		print len(sys.argv)
		print "fail"
elif "alivemeta" in path and "get" in op:
	if data == server_check[check]:
		print "pass"
	elif c2 and data == server_check[check2]:
		print "pass"
	else:
		print "zkdata:", data
		print "checkdata:", server_check[check]
		if c2:
			print "checkdata2:", server_check[check2]
		print len(sys.argv)
		print "fail"
elif "alivemeta" in path and "ls" in op:
	data = []
	for l in out.split("\n"):
		if l != "":
			data.append(l)
	if set(data) == set(alive_check[check]):
		print "pass"
	else:
		print "zkdata:", data
		print "checkdata:", alive_check[check]
		print "fail"
