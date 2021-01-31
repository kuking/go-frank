### AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)

(disk is encrypted, see below max IO bandwidth)

```
# 1000 million events of 500 bytes, 500GB of storage used
% ./franki -ps 1024 -miop 1000 -evs 500 pub_bench
Totals=1000M IOP; 500000MB; Perfs=1.82M IOPS; 909.29MB/s; avg 549ns/iop; [100%]     
% ./franki sub_bench
Totals=1000M IOP; 500000MB; Perfs=1.72M IOPS; 862.37MB/s; avg 579ns/iop; [+Inf%]  
```

IO saturated, disk is encrypted, see below:

```
% head -c 100000000000 /dev/zero | pv -ba > file
93.1GiB [ 831MiB/s]
% cat file | pv -ba >/dev/null
93.1GiB [ 878MiB/s]
```

```
# 1000 million events of 100 bytes, 100GB of storage used
% ./franki -ps 1024 -miop 1000 pub_bench
Totals=1000M IOP; 100000MB; Perfs=5.73M IOPS; 572.82MB/s; avg 174ns/iop; [100%]     
 % ./franki sub_bench                    
Totals=1000M IOP; 100000MB; Perfs=6.27M IOPS; 627.00MB/s; avg 159ns/iop; [+Inf%] 
```

prob. IO saturated

```
# 1000 million events of 10 bytes, 10GB of storage used
% ./franki -ps 1024 -miop 1000 -evs 10 pub_bench
Totals=1000M IOP; 10000MB; Perfs=10.75M IOPS; 107.50MB/s; avg 93ns/iop; [100%]     
% ./franki sub_bench                            
Totals=1000M IOP; 10000MB; Perfs=8.48M IOPS; 84.76MB/s; avg 117ns/iop; [+Inf%]     
```

CAS saturated, see below a test with 0 bytes elements (just headers)

```
% ./franki -evs 0 pub_bench
Totals=100M IOP; 0MB; Perfs=11.71M IOPS; 0.00MB/s; avg 85ns/iop; [100%]     
% ./franki sub_bench
Totals=100M IOP; 0MB; Perfs=8.38M IOPS; 0.00MB/s; avg 119ns/iop; [+Inf%]
```

```
% ./franki pub_bench
Totals=100M IOP; 10000MB; Perfs=7.13M IOPS; 713.13MB/s; avg 140ns/iop; [100%]    
% ./franki sub_bench
Totals=100M IOP; 10000MB; Perfs=7.98M IOPS; 798.19MB/s; avg 125ns/iop; [+Inf%]
```
Default test parameters

### Raspberry PI 4

SD card can't handle more than 11MB/s write, 39MB/s read.

```
# 100 bytes records, 100M, Raspberry Pi 400 Rev 1.0, 4GB RAM, SD CARD
$ ./franki pub_bench
Totals=100M IOP; 10000MB; Perfs=0.21M IOPS; 21.40MB/s; avg 4.672µs/iop; [100%]
$ ./franki sub_bench
Totals=100M IOP; 10000MB; Perfs=0.40M IOPS; 40.46MB/s; avg 2.471µs/iop; [+Inf%]
```


### Intel(R) Core(TM) i7-8559U CPU @ 2.70GHz (2018 mac 1TB NVMe)

```
$  head -c 100000000000 /dev/zero | pv -ba > file
93.1GiB [2.02GiB/s]
$ cat file | pv -ba >/dev/null
93.1GiB [2.17GiB/s]
```

```
$ ./franki -ps 1024 -miop 1000 pub_bench
Totals=1000M IOP; 100000MB; Perfs=6.21M IOPS; 621.36MB/s; avg 160ns/iop; [100%]
$ ./franki sub_bench
Totals=1000M IOP; 100000MB; Perfs=2.01M IOPS; 200.59MB/s; avg 498ns/iop; [+Inf%]

```

```
$ ./franki -evs 0 pub_bench
Totals=100M IOP; 0MB; Perfs=12.34M IOPS; 0.00MB/s; avg 81ns/iop; [100%]
$ ./franki sub_bench
Totals=100M IOP; 0MB; Perfs=10.53M IOPS; 0.00MB/s; avg 95ns/iop; [+Inf%]
```
