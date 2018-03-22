#!/bin/sh
go get -u golang.org/x/vgo &> goget.log
$GOPATH/bin/vgo build -o bin/dreamer &> build.log

pkill dreamer
cp bin/dreamer $HOME/services/
cd $HOME/services
nohup ./dreamer &> dreamer.log &
