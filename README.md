# iclangscripts
Useful scripts for IClang.

(1) front_back_time.cpp

```shell
g++ front_back_time.cpp -o front_back_time
./front_back_time <dir>
```

Recursively traverse all *.iclang/compile.txt under `dir`, calculate the total front-end time and the total back-end time.

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
