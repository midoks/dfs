#! /bin/bash

ROOT_DIR=`pwd`

rm -rf dfs $ROOT_DIR/data/dfs.db
rm -rf dfs $ROOT_DIR/files/*


ps -ef|grep dfs | grep -v grep  | grep -v godfs | awk '{print $2}' | xargs kill

go build *.go

mkdir -p $ROOT_DIR/tmp/test1
cp dfs $ROOT_DIR/tmp/test1
rm -rf $ROOT_DIR/tmp/test1/data/dfs.db
rm -rf $ROOT_DIR/tmp/test1/files/*



mkdir -p $ROOT_DIR/tmp/test2
cp dfs $ROOT_DIR/tmp/test2
rm -rf  $ROOT_DIR/tmp/test2/data/dfs.db
rm -rf $ROOT_DIR/tmp/test2/files/*

rm -rf dfs

cd $ROOT_DIR/tmp/test1 && ./dfs &
cd $ROOT_DIR/tmp/test2 && ./dfs
