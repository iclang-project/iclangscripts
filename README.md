# iclangscripts
Useful scripts for IClang.

Build:

```shell
./build.sh
```

(1) sta

```shell
sta <dir> <baseTsMs>
```

Recursively traverse all *.iclang under `dir` whose `EndTsMs(in compile.json) < `, accumulate `IClangDirStat`.

(2) 2x

```shell
2x <benchmarkdir> <projects> <scriptname> <logdir> <enableIClang>
# For example: ./2x ../ all checkbugs ./log 1
# Note: (1) <projects> can be 'all', or your projects separated by ':'. For example: llvm:cpython
#       (2) Do not provide '.sh' in <scriptname>
```

Run `<scriptname>`.sh through a coroutine pool of size 2 in:

```shell
<benchmarkdir>/llvm
<benchmarkdir>/cvc5
<benchmarkdir>/z3
<benchmarkdir>/sqlite
<benchmarkdir>/cpython
<benchmarkdir>/postgres
```

Take llvm as an example, you can use `tail -f <logdir>/llvm/<scriptname>.log` to see the log.

(3) 2x_100

```shell
2x_100 <benchmarkdir> <projects> <logdir> <enableIClang>
# For example: ./2x_100 ../ all ./log 1
```

Run 100 commits through a coroutine pool of size 2 in:

```shell
<benchmarkdir>/llvm
<benchmarkdir>/cvc5
<benchmarkdir>/z3
<benchmarkdir>/sqlite
<benchmarkdir>/cpython
<benchmarkdir>/postgres
```

Take llvm as an example, you can use `tail -f <logdir>/llvm/100commits.log` to see the log,
and you can use `cat <logdir>/llvm/100commits.json` to see the json result.