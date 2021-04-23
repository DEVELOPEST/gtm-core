# Rewrite test

old = []
with open('./old.txt') as f:
    old = [line.rstrip() for line in f]
new = []
with open('./new.txt') as f:
    new = [line.rstrip() for line in f]

total = [f"{a} {b}" for (a,b) in zip(old, new)]

for commit in total:
    print(commit)