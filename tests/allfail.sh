#!/bin/bash
suite=true

# build source
make || exit 1
echo

echo "basic fail"
. tests/basicfail.sh
: '
while true; do
    read -p "Continue? [y/n] " yn
    case $yn in
        [Yy]* ) 
		break	
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
'
echo "lead fail"
. tests/leadcoordfail.sh
: '
while true; do
    read -p "Continue? [y/n] " yn
    case $yn in
        [Yy]* ) 
		break
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
'
echo "fast fail"
. tests/fastfail.sh
: '
while true; do
    read -p "Continue? [y/n] " yn
    case $yn in
        [Yy]* ) 
		break
        ;;
        * ) 
        echo "Please answer yes."
        ;;
    esac
done
'
