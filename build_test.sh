#! /bin/bash

ROOT_DIR=`pwd`

mkdir -p $ROOT_DIR/tmp/test1

go build dfs.go

mv dfs $ROOT_DIR/tmp/test1

cd $ROOT_DIR/tmp/test1 && ./dfs
