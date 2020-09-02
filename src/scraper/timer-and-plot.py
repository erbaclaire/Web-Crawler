import pandas as pd
import numpy as np 
import matplotlib.pyplot as plt
from datetime import datetime
import math
import subprocess
import sys
import os

# declare global vars
problems = ["genres1", "genres2", "genres3"]
threads = [1, 2, 4, 6, 8]
seq = []
trials = 20
if len(sys.argv) == 2:
	trials = int(sys.argv[1])

df = pd.DataFrame(np.array(threads), columns=['threads'], dtype=int)

# calculate parallel and sequential times of execution
for p in problems:

	file = "./in-files/" + str(p) + ".txt"
	out_file = "./out-files/" + str(p) + "_Out.txt"
	check_file = "./out-files/answers/" + str(p) + "_Out_Ans.txt"

	# parallel version
	arr = np.empty(0)
	for n in threads:
		timeTotal = 0
		for i in range(trials): # run 20 times for each problem and block size
			with open(file, "r") as f:
				with open(out_file, "w") as f2:
					start = datetime.now()
					subprocess.run(args=['go', 'run', 'scraper.go', "-p="+str(n)], stdin=f, stdout=f2, check=True)
					end = datetime.now()
			timeTotal += (end-start).total_seconds()
			# compare to make sure matches answer
			subprocess.run(args=['go', 'run', 'check.go', out_file, check_file], check=True)
			# print that done
			print("DONE: problem:", p, ", number of threads:", n, ", iteration:", i)
		arr = np.append(arr, timeTotal/trials)
	arr_df = pd.DataFrame(arr, columns=[str(p)])
	df = pd.concat([df, arr_df], axis=1)

	# sequential version
	for i in range(trials): # run 20 times for each problem
		with open(file, "r") as f:
			with open(out_file, "w") as f2:
				start = datetime.now()
				subprocess.run(args=['go', 'run', 'scraper.go'], stdin=f, stdout=f2, check=True)
				end = datetime.now()
		timeTotal += (end-start).total_seconds()
		# compare to make sure matches answer
		subprocess.run(args=['go', 'run', 'check.go', out_file, check_file], check=True)
		# print that done
		print("DONE: problem: ", p, "- sequential , iteration:", i)
	seq.append(timeTotal/trials)

print(df)
print(seq)

# calculate speedup
for i, val in enumerate(problems):
	df[val] = seq[i]/df[val]

print(df)

# plot
ax = plt.gca()
df.plot(kind='line', x='threads', y='genres1', marker='D', ax=ax)
df.plot(kind='line', x='threads', y='genres2', marker='D', ax=ax)
df.plot(kind='line', x='threads', y='genres3', marker='D', ax=ax)
plt.title("Number of Threads v. Speedup")
plt.ylabel("Speedup")
plt.xlabel("Number of Threads (N)")
plt.legend()
ax.yaxis.grid()
plt.show()
# plt.savefig('benchmark.png')
