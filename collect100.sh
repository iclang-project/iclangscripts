set -e
set -x
if [ -z "$1" ]; then
	echo "Please provide a project path"
	exit 1
fi

if [ -z "$2" ]; then
	echo "Please provide a log path"
	exit 1
fi

logPath=$(readlink -f $2)
rm -f $logPath

export ICLANG=1
export INSTALLFLAG=0

start_time=$(date +%s)

cd $1
./config.sh
./build.sh

okNum=0
errNum=0

while true; do
    if ((okNum == 100)) || ((errNum == 100)); then
        break
    fi
    find . -type f -name "diff.txt" -path "*.iclang/*" | xargs -I {} rm "{}"
    cd src
    git checkout HEAD^
    commitId=$(git rev-parse HEAD)
    cd ..
    set +e
    start_time_t=$(date +%s)
    ./build.sh
    if [ $? -ne 0 ]; then
	errNum=$[$errNum+1]
	echo "$commitId error" >> $logPath
	continue
    fi
    end_time_t=$(date +%s)
    diff_time_t=$((end_time_t-start_time_t))
    set -e
    diffNum=$(find . -type f -name "diff.txt" -path "*.iclang/*" | wc -l)
    if [ "$diffNum" -eq 0 ]; then
    	echo "$commitId no" >> $logPath
    else
	okNum=$[$okNum+1]
	echo "$commitId yes $diff_time_t" >> $logPath
    fi
done

end_time=$(date +%s)
diff_time=$((end_time-start_time))
echo "done, time=$diff_time"
