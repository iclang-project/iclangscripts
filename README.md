# iclangscripts
Useful scripts for IClang.

Build:

```shell
./build.sh
```

(1) sta

```shell
sta <dir> <base-timestamp-ms>
```

Recursively traverse all *.iclang under `dir`, collect the following data:

```shell
FileNum       
CompileTimeMs 
FrontTimeMs   
BackTimeMs    
FileSizeB     
SrcLoc        
PPLoc         
FuncNum       
FuncLoc       
FuncXNum      
FuncXLoc      
FuncZNum      
FuncZLoc      
FuncVNum      
FuncVLoc
```

Note that we only consider the iclang dir whose `endTimestampMs` >= `<base-timestamp-ms>`.

(2) collect100.sh / collect100_fossil.sh

```shell
./collect100.sh <git-project-path> <log-path>
```

cd `git-project-path`, starting from HEAD, collect 100 commits with code changes from new to old, save the result in `log-path`, format:

```shell
# each line, from new to old:
commitId yes|no|error [time(s)]
# yes: commit with code changes
# no: commit without code changes
# error: compilation error
# time: build time(s) for 'yes' and 'no'
# exit when the total number of 'error' >= 100
```

`collect100_fossil.sh` is the same as `collect100.sh`, it works for fossil projects.

(3) collect100_cmp.sh / collect100_fossil_cmp.sh

```shell
./collect100.sh <git-project-path> <log-path> <commit-num>
```

cd `git-project-path`, starting from HEAD, collect `<commit-num>` commits from new to old, save the result in `log-path`, format:

 ```shell
 # each line, from new to old:
 commitId yes|error [time(s)]
 ```

Unlike`collect100.sh`, `collect100_cmp.sh` does not enable IClang, but rather enable a normal compiler.

We will leverage the result of  `collect100_cmp.sh` to check the result of`collect100.sh`.

(4) format_commit_log.sh

```shell
./format_commits_log.sh <original-commits-log> <formated-commits-log>
```

Just change each `commitId yes|no|error [time(s)]` to `commitId yes|no|error`.

(5) 2x

```shell
2x <benchmarkdir> <scriptname> <logdir>
# Note: Do not provide '.sh' in <scriptname>
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

(6) 2x_100

```shell
2x_100 <benchmarkdir> <logdir>
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