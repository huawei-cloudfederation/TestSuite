import os
import sys
import json


from pprint import pprint

with open('fed.conf') as data_file:
    data = json.load(data_file)

print "data : ", data["DC LIST"]["DC1"]
ip=data["DC LIST"]["DC1"]
os.system("scp -i test.pem mesos_fed.sh fed_mod.sh ubuntu@%(ip)s:~"% locals())
os.system("ssh -i test.pem ubuntu@%(ip)s './mesos_fed.sh'"% locals())
os.system("ssh -i test.pem ubuntu@%(ip)s './fed_mod.sh'"% locals())
