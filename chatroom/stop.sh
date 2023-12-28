#!/bin/bash

PIDS=`ps -ef | grep chatroom | grep -v grep | awk '{print $2}'`
if [ "x${PIDS}" != "x" ];then
	echo $PIDS | xargs kill -9  
fi

