suite=true

# build source
make || exit 1
echo

. tests/read.sh
. tests/write.sh
. tests/unlink.sh
. tests/vim.sh
. tests/deepdir.sh
. tests/zkmeta.sh
data_mvmt
. tests/basicfail.sh
. tests/leadcoordfail.sh
. tests/fastfail.sh
