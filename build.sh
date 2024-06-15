set -e
set -x

go build -o 2x ./src_2x
go build -o 2x_100 ./src_2x_100
go build -o sta ./src_sta
go build -o changes100 ./src_changes100
go build -o patch ./src_patch
go build -o captor ./src_captor