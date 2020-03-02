#!/user/bin/env python3

#
# this file takes output of Yelp's secret-detect from standard input and process it
#
# detect-secrets scan | python3 ./scripts/secret-post-process.py ; echo $?
#

import json
import sys
import os

whiteList = ["go.sum", "src/unit-test/example_private_key"]

stdin=''

for line in sys.stdin:
  if line == "\n":
    lb += 1
    if lb == 2:
        break
  else:
    lb = 0
    stdin += line

results = json.loads(stdin)

foundGoSum = False
error = False
for (k, v) in results["results"].items():
    if k == "go.sum":
        foundGoSum = True
    if k not in whiteList:
        print(k, v)
        error = True

if not foundGoSum:
    print("Error: fail to process expected go.sum")
    os._exit(2)

if error:
    print("Error: above secret detected")
    os._exit(3)
else:
    print("successful")
