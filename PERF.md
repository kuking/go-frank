
All the tests run without noticiable CPU usage. It is an IO task.

```
# 150 bytes records, 1B, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki -ps 1024 pub_bench
Total=1000M IOP; 150000Mb Bytes. Performance=4.12M IOPS; 618.30Mb/s.
% ./franki -ps 1024 sub_bench
Total=1000M IOP; 150000Mb Bytes. Performance=4.61M IOPS; 691.81Mb/s.
```

```
# 100 bytes records, 1B, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki -ps 1024 pub_bench
Total=1000M IOP; 100000Mb Bytes. Performance=5.54M IOPS; 553.68Mb/s.    
% ./franki sub_bench
Total=1000M IOP; 100000Mb Bytes. Performance=5.79M IOPS; 579.33Mb/s.
```

```
# 100 bytes records, 100M, AMD Ryzen 7 3800X, 64GB RAM, Force MP510 (ECFM22.5)
% ./franki pub_bench    
Total= 100M IOP; 10000Mb Bytes. Performance=6.09M IOPS; 608.51Mb/s.    
% ./franki sub_bench
Total= 100M IOP; 10000Mb Bytes. Performance=8.12M IOPS; 811.55Mb/s. 
```

```
# 100 bytes records, 100M, Raspberry Pi 400 Rev 1.0, 4GB RAM, SD CARD
$ ./franki pub_bench
Total= 100M IOP; 10000Mb Bytes. Performance=0.12M IOPS; 11.87Mb/s.
$ ./franki sub_bench
Total= 100M IOP; 10000Mb Bytes. Performance=0.39M IOPS; 39.03Mb/s.
```
