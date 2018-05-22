#!/bin/bash
#install go
sudo apt-get update
sudo apt-get install golang-go openjdk-8-jre-headless -y 
mkdir $HOME/go

#install zookeeper
wget http://mirrors.gigenet.com/apache/zookeeper/zookeeper-3.4.12/zookeeper-3.4.12.tar.gz
tar -xvzf zookeeper-3.4.12.tar.gz
rm zookeeper-3.4.12.tar.gz
cp zookeeper-3.4.12/conf/zoo_sample.cfg zookeeper-3.4.12/conf/zoo.cfg
mkdir zkdata
sed -i 's/dataDir=\/tmp\/zookeeper/dataDir=\/home\/ubuntu\/zkdata/' zookeeper-3.4.12/conf/zoo.cfg

#setup directories and bash files
mkdir servermount
mkdir clientmount
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> .profile
echo "export GOPATH=$HOME/go" >> .profile
touch .bash_aliases
echo "alias runserver=\"fs-server $HOME/servermount\"" >> .bash_aliases
echo "alias runclient=\"fs-front $HOME/clientmount\"" >> .bash_aliases
echo "alias proj=\"cd $HOME/go/src/proj\"" >> .bash_aliases
source ~/.bash_aliases
source ~/.profile
source ~/.bashrc

# prep source code
go get github.com/hanwen/go-fuse/...
go get github.com/samuel/go-zookeeper/...
git clone https://github.com/kkuo42/CSE223B-FinalProject.git $HOME/go/src/proj
