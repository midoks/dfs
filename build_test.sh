#! /bin/bash

ROOT_DIR=`pwd`

rm -rf dfs $ROOT_DIR/data/dfs.db
rm -rf dfs $ROOT_DIR/files/*


mkdir -p $ROOT_DIR/tmp/test1

go build dfs.go

mv dfs $ROOT_DIR/tmp/test1
rm -rf dfs $ROOT_DIR/tmp/test1/data/dfs.db
rm -rf dfs $ROOT_DIR/tmp/test1/files/*


cd $ROOT_DIR/tmp/test1 && ./dfs
