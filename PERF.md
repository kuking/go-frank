
```
# 500 bytes records, 1B, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki -ps 1024 -miop 1000 -evs 500 pub_bench
Totals=1000M IOP; 500000MB; Perfs=1.39M IOPS; 692.52MB/s; avg 721ns/iop; [100%]     
% ./franki sub_bench                             
Totals=1000M IOP; 500000MB; Perfs=1.71M IOPS; 856.40MB/s; avg 583ns/iop; [+Inf%]  
```
prob. IO saturated, disk is encrypted, see below:

```
# AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% head -c 100000000000 /dev/zero | pv -ba > file
93.1GiB [ 836MiB/s]
```

```
# 100 bytes records, 1B, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki -ps 1024 -miop 1000 pub_bench
Totals=1000M IOP; 100000MB; Perfs=5.73M IOPS; 572.82MB/s; avg 174ns/iop; [100%]     
 % ./franki sub_bench                    
Totals=1000M IOP; 100000MB; Perfs=6.27M IOPS; 627.00MB/s; avg 159ns/iop; [+Inf%] 
```
prob. IO saturated

```
# 10 bytes records, 1B, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki -ps 1024 -miop 1000 -evs 10 pub_bench
Totals=1000M IOP; 10000MB; Perfs=10.75M IOPS; 107.50MB/s; avg 93ns/iop; [100%]     
% ./franki sub_bench                            
Totals=1000M IOP; 10000MB; Perfs=8.48M IOPS; 84.76MB/s; avg 117ns/iop; [+Inf%]     
```
CAS saturated

```
# 100 bytes records, 100M, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki  pub_bench 
Totals=100M IOP; 10000MB; Perfs=6.97M IOPS; 697.38MB/s; avg 143ns/iop; [100%]     
% ./franki sub_bench 
Totals=100M IOP; 10000MB; Perfs=8.07M IOPS; 807.09MB/s; avg 123ns/iop; [+Inf%]     
```
Stream fits in memory

```
# 100 bytes records, 100M, Raspberry Pi 400 Rev 1.0, 4GB RAM, SD CARD
$ ./franki pub_bench
Total= 100M IOP; 10000Mb Bytes. Performance=0.12M IOPS; 11.87Mb/s.
$ ./franki sub_bench
Total= 100M IOP; 10000Mb Bytes. Performance=0.39M IOPS; 39.03Mb/s.
```
