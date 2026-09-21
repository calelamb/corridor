# Authored Go packages only; generated adapters are validated by contracts/integration.
NR == 1 { next }
/\/api\/generated.go:/ || /\/db\/generated\// { next }
{
 name=$1; sub(/\/[^\/]+:[^:]+$/, "", name)
 total[name]+=$2
 if ($3>0) covered[name]+=$2
}
END {
 failed=0
 for (name in total) {
  percent=100*covered[name]/total[name]
  printf "%s %.1f%%\n", name, percent
  if (percent<80) failed=1
 }
 if (length(total)==0) failed=1
 exit failed
}
