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

if [ -z "$3" ]; then
	echo "Please provide the total number of commits"
	exit 1
fi

logPath=$(readlink -f $2)
rm -f $logPath

export INSTALLFLAG=0

start_time=$(date +%s)

cd $1
./config.sh
./build.sh

curNum=0
totalNum=$3

while true; do
    if ((curNum == totalNum)); then
        break
    fi
    curNum=$[$curNum+1]
    find . -type f -name "diff.txt" -path "*.iclang/*" | xargs -I {} rm "{}"
    cd src
    fossil update prev
    commitId=$(fossil info | awk '/checkout:/ {print $2}')
    cd ..
    set +e
    start_time_t=$(date +%s)
    ./build.sh
    if [ $? -ne 0 ]; then
	echo "$commitId error" >> $logPath
	continue
    fi
    end_time_t=$(date +%s)
    diff_time_t=$((end_time_t-start_time_t))
    set -e
    diffNum=$(find . -type f -name "diff.txt" -path "*.iclang/*" | wc -l)
    echo "$commitId yes $diff_time_t" >> $logPath
done

end_time=$(date +%s)
diff_time=$((end_time-start_time))
echo "done, time=$diff_time"
