set -e
if [ -z $1 ]; then
	echo "Please provide original commits log"
	exit 1
fi

if [ -z $1 ]; then
	echo "Please provide output path"
	exit 1
fi

rm -f $2
cat $1 | awk '{print $1, $2}' > $2
