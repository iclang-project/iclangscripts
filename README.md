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

Recursively traverse all *.iclang under `dir` whose `EndTsMs(in compile.json) < `, accumulate the following data:

```shell
{
    "compStat": {
        "originalCommand": "",
        "ppCommand": "",
        "compileCommand": "",
        "inputAbsPath": "",
        "outputAbsPath": "",
        "startTsMs": 0,
        "oldCDGStat": {
            "flag": 0,
            "funcNum": 0,
            "funcDefNum": 0,
            "whiteFuncNum": 0,
            "specFuncNum": 0,
            "funcZReason": {
                "inlineNum": 0,
                "invalidFuncNum": 0,
                "unknownDepNum": 0,
                "depNum": 0
            },
            "inlineEdgeNum": 0,
            "startTsMs": 0,
            "endTsMs": 0,
            "doaElemNum": [],
            "modifiedElemNum": [],
            "edgesNum": []
        },
        "newCDGStat": {
            "flag": 0,
            "funcNum": 0,
            "funcDefNum": 0,
            "whiteFuncNum": 0,
            "specFuncNum": 0,
            "funcZReason": {
                "inlineNum": 0,
                "invalidFuncNum": 0,
                "unknownDepNum": 0,
                "depNum": 0
            },
            "inlineEdgeNum": 0,
            "startTsMs": 0,
            "endTsMs": 0,
            "doaElemNum": [],
            "modifiedElemNum": [],
            "edgesNum": []
        },
        "startLinkTsMs": 0,
        "endTsMs": 0
    },
    "compileTimeMs": 0,
    "frontTimeMs": 0,
    "backTimeMs": 0,
    "ppTimeMs": 0,
    "oldCDGTimeMs": 0,
    "cc1FrontTimeMs": 0,
    "newCDGTimeMs": 0,
    "cc1BackTimeMs": 0,
    "linkTimeMs": 0,
    "fileNum": 0,
    "fileSizeB": 0,
    "srcLoc": 0,
    "ppLoc": 0
}
```

(2) 2x

```shell
2x <benchmarkdir> <projects> <scriptname> <logdir>
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
2x_100 <benchmarkdir> <projects> <logdir>
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

(4) collect100.sh / collect100_fossil.sh

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

(5) collect100_cmp.sh / collect100_fossil_cmp.sh

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

(6) format_commit_log.sh

```shell
./format_commits_log.sh <original-commits-log> <formated-commits-log>
```

Just change each `commitId yes|no|error [time(s)]` to `commitId yes|no|error`.