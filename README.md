# go-frank streaming framework

Two flavours: in-memory and persistent streams, multiple-producers, multiple-consumers; raw -fast- stream replicator.

It is fast (see [PERF.md](PERF.md)), it won't use the typical mutexes, locks, channels, etc. It incorporates ideas
from: [Aeron](https://github.com/real-logic/Aeron),
and [Mechanical Sympathy blog](https://mechanical-sympathy.blogspot.com).

It tries to find a balance between maximum possible performance, which comes at high cost i.e. flyweights,
non-allocation, buffer pools, etc. and a practical simple framework that performs very fast with low-latency.

**Status:** Mostly works, you are welcome to participate.

## Fluid syntax

Extracts from [demo1](cli/demo1/main.go):

```go
func textFile() {
    lines := frank.TextFileStream("README.md").Count()
    chars := frank.TextFileStream("README.md").
        Map(func (line string) int { return len(line) + 1 }).
        Sum().
        First()
    fmt.Printf("README.md has %v lines and %v characters.\n", lines, chars)

    title := frank.TextFileStream("README.md").
    Filter(func (s string) bool { return len(s) < 1 || s[0] != '#' }).
        Map(func (s string) string { return strings.TrimSpace(s[1:]) }).
        First()
    fmt.Printf("README.md title is: %v\n", title)
}
```

```
README.md has 94 lines and 3785 characters.
README.md title is: go-frank streaming framework
```

## Performance

Extracts from [PERF.md](PERF.md), Total data is 500GiB, which won't fit into main memory, disk is encrypted (lower
performance), with 500 bytes events, it averages 1.82M inserts-per-second (throughput 909MiB/s), and 1.72M
reads-per-second (862MiB/s). Multi-producer/thread safe.

```
# 1000 million events of 500 bytes, 500GB of storage used
% ./franki -ps 1024 -miop 1000 -evs 500 pub_bench
Totals=1000M IOP; 500000MB; Perfs=1.82M IOPS; 909.29MB/s; avg 549ns/iop; [100%]     
% ./franki sub_bench
Totals=1000M IOP; 500000MB; Perfs=1.72M IOPS; 862.37MB/s; avg 579ns/iop; [+Inf%]  
```

## Replication

It is a work-in-progress, but over 10GbE with OK hardware peaks at 285MiB/s (~3Gbits) transfers.

### Sender

```
% ./frankr send r@persistent-stream host.local:1234
so sending: r @ persistent-stream -> host.local:1234
[0: R: 575897886 (274.00MiB/s) 5.43% W: 10600003978 (0.00MiB/s)]
[0: R: 863670286 (274.00MiB/s) 8.15% W: 10600003978 (0.00MiB/s)]
[0: R: 1147062448 (270.00MiB/s) 10.82% W: 10600003978 (0.00MiB/s)]
[0: R: 1432128456 (271.00MiB/s) 13.51% W: 10600003978 (0.00MiB/s)]
[0: R: 1722462134 (276.00MiB/s) 16.25% W: 10600003978 (0.00MiB/s)]
[0: R: 2004264614 (268.00MiB/s) 18.91% W: 10600003978 (0.00MiB/s)]
[0: R: 2288827758 (271.00MiB/s) 21.59% W: 10600003978 (0.00MiB/s)]
[0: R: 2571388032 (269.00MiB/s) 24.26% W: 10600003978 (0.00MiB/s)]
[0: R: 2860003662 (275.00MiB/s) 26.98% W: 10600003978 (0.00MiB/s)]
[0: R: 3141271266 (268.00MiB/s) 29.63% W: 10600003978 (0.00MiB/s)]
[0: R: 3436926886 (281.00MiB/s) 32.42% W: 10600003978 (0.00MiB/s)]
[0: R: 3732372096 (281.00MiB/s) 35.21% W: 10600003978 (0.00MiB/s)]
```

### Receiver

```
% ./frankr accept streams/ :1234
2021/03/14 10:41:38 Accepting: :1234; streams in: streams/ (accepting -not implemented- [*])
[0: R: 0 (0.00MiB/s) 0.00% W: 432876436 (271.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 729910268 (283.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 1009488020 (266.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 1294961386 (272.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 1581864586 (273.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 1859224500 (264.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 2147483754 (274.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 2421702570 (261.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 2710795836 (275.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 2994601716 (270.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 3282133602 (274.00MiB/s)]
[0: R: 0 (0.00MiB/s) 0.00% W: 3582852100 (286.00MiB/s)]
```