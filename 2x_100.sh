echo "Disable IClang =========="
./2x_100 ../ all ./clanglog
echo "Enable IClang =========="
ICLANG=1 ./2x_100 ../ all ./iclanglog