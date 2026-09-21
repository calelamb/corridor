# Authored Go packages; deduplicate blocks when -coverpkg instruments callers.
NR == 1 { next }
/\/api\/generated.go:/ || /\/db\/generated\// { next }
{
 block=$1
 statements[block]=$2
 if ($3>0) hits[block]=1
}
END {
 for (block in statements) {
  name=block; sub(/\/[^\/]+:[^:]+$/, "", name)
  total[name]+=statements[block]
  if (hits[block]) covered[name]+=statements[block]
 }
 failed=0
 for (name in total) {
  percent=100*covered[name]/total[name]
  printf "%s %.1f%%\n", name, percent
  if (percent<80) failed=1
 }
 if (length(total)==0) failed=1
 exit failed
}
