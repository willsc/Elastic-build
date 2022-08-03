```

sh-5.0$ grep -v "rem_address" /proc/net/tcp | awk 'function hextonum(str, ret, n, i, k, c) {if (str ~ /^0[xX][0-9a-FA-F]+$/) {str = substr(str, 3);n = length(str);ret = 0;for (i = 1; i <= n; i++) {c = substr(str, i, 1);c = tolower(c);k = index("123456789abcdef", c);ret = ret * 16 + k}} else ret = "NOT-A-NUMBER";return ret} {y=hextonum("0x"substr($2,index($2,":")-2,2));x=hextonum("0x"substr($3,index($3,":")-2,2));for (i=5; i>0; i-=2) {x = x"."hextonum("0x"substr($3,i,2));y = y"."hextonum("0x"substr($2,i,2));} print y":"hextonum("0x"substr($2,index($2,":")+1,4))" "x":"hextonum("0x"substr($3,index($3,":")+1,4));}'
0.0.0.0:9200 0.0.0.0:0
0.0.0.0:9300 0.0.0.0:0
10.0.2.70:60530 10.0.3.96:9300
10.0.2.70:9300 10.0.3.162:54484
10.0.2.70:9300 10.0.3.162:54478
10.0.2.70:9300 10.0.3.162:54446
127.0.0.1:36852 127.0.0.1:9200
10.0.2.70:60924 10.0.3.162:9300
10.0.2.70:9300 10.0.3.96:43598
10.0.2.70:60540 10.0.3.96:9300
127.0.0.1:38376 127.0.0.1:9200
10.0.2.70:60568 10.0.3.96:9300
10.0.2.70:9300 10.0.3.96:43662
10.0.2.70:9300 10.0.3.96:43654
10.0.2.70:9300 10.0.3.162:54434
10.0.2.70:9300 10.0.3.162:54462
127.0.0.1:46092 127.0.0.1:9200
10.0.2.70:60588 10.0.3.96:9300
10.0.2.70:60554 10.0.3.96:9300
10.0.2.70:60498 10.0.3.96:9300
10.0.2.70:9300 10.0.3.96:43602
10.0.2.70:60552 10.0.3.96:9300
10.0.2.70:9300 10.0.3.96:43628
10.0.2.70:60916 10.0.3.162:9300
127.0.0.1:48306 127.0.0.1:9200
10.0.2.70:9300 10.0.3.162:54414
10.0.2.70:60502 10.0.3.96:9300
10.0.2.70:9300 10.0.3.162:54412
10.0.2.70:9300 10.0.3.96:43580
10.0.2.70:60576 10.0.3.96:9300
10.0.2.70:9300 10.0.3.96:43642
10.0.2.70:9300 10.0.3.96:43584
10.0.2.70:60900 10.0.3.162:9300
10.0.2.70:60928 10.0.3.162:9300
10.0.2.70:60892 10.0.3.162:9300
10.0.2.70:60846 10.0.3.162:9300
10.0.2.70:9300 10.0.3.96:43666
10.0.2.70:9300 10.0.3.162:54504
10.0.2.70:60876 10.0.3.162:9300
10.0.2.70:9300 10.0.3.162:54492
10.0.2.70:60594 10.0.3.96:9300
10.0.2.70:9300 10.0.3.96:43644
127.0.0.1:34270 127.0.0.1:9200
127.0.0.1:35108 127.0.0.1:9200
10.0.2.70:60514 10.0.3.96:9300
10.0.2.70:60884 10.0.3.162:9300
10.0.2.70:60854 10.0.3.162:9300
10.0.2.70:60866 10.0.3.162:9300
10.0.2.70:9300 10.0.3.162:54486
10.0.2.70:9300 10.0.3.162:54430
10.0.2.70:60902 10.0.3.162:9300
10.0.2.70:9300 10.0.3.96:43616
```


```
declare -a array=($(tail -n +2 /proc/net/tcp | cut -d":" -f"3"|cut -d" " -f"1")) && for port in ${array[@]}; do echo $((0x$port)); done

sh-5.0$ declare -a array=($(tail -n +2 /proc/net/tcp | cut -d":" -f"3"|cut -d" " -f"1")) && for port in ${array[@]}; do echo $((0x$port)); done
9200
9300
40892
40878
43616
40850
43584
9300
9300
40828
9300
40908
37494
52712
9300
9300
9300
37212
43628
57838
41304
9300
9300
9300
43644
43666
9300
40920
40932
9300
9300
43654
9300
9300
43642
60550
40864
43580
9300
9300
9300
40946
43662
9300
40930
43602
40838
9300
9300
9300
9300
43598
```


```
declare -a open_ports=($(cat /proc/net/tcp /proc/net/raw /proc/net/udp | grep -v "local_address" | awk '{ print $2 }'))

# Define function for converting
dec2ip () {
    ip=$1
    s=""
    for i in {1..4}; do 
    s='.'$((ip%256))$s && ((ip>>=8)); 
    done; 
    echo ${s:1} | sed 's/\./\n/g' | tac | sed ':a; $!{N;ba};s/\n/./g'
}

# Show all open ports and decode hex to dec
for tuple in ${open_ports[*]}; do 
    port=${tuple#*:}
    ip=${tuple%:*}
    echo $(dec2ip $((0x${ip}))):$((0x${port})); 
done | sort | uniq -c



```


```
sh-5.0$ PORT=9300;find /proc -lname "socket:\[$(cat /proc/net/* | awk -F " " '{print $2 ":" $10 }' | grep -i `printf "%x:" $PORT` | head -n 1 | awk -F ":" '{print $3}')\]" 2> /dev/null | head -n 1 | awk -F "/" '{print "PID="$3}'
cat: /proc/net/dev_snmp6: Is a directory
cat: /proc/net/netfilter: Is a directory
cat: /proc/net/rpc: Is a directory
cat: /proc/net/stat: Is a directory
PID=8
sh-5.0$ ps -ef
UID        PID  PPID  C STIME TTY          TIME CMD
elastic+     1     0  0 18:42 ?        00:00:00 /bin/tini -- /usr/local/bin/docker-entrypoint.sh eswrapper
elastic+     8     1  2 18:42 ?        00:00:30 /usr/share/elasticsearch/jdk/bin/java -Xshare:auto -Des.networkaddress.cache.ttl=60 -Des.networkaddress.cache.negative.ttl=10 -XX:+Always
elastic+   180     8  0 18:42 ?        00:00:00 /usr/share/elasticsearch/modules/x-pack-ml/platform/linux-x86_64/bin/controller
elastic+   878     0  0 18:56 pts/0    00:00:00 sh
elastic+  1354   878  0 19:00 pts/0    00:00:00 ps -ef

```
