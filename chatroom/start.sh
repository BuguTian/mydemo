#!/bin/bash

./chatroom -daemon -forever > /dev/null 2>&1
sleep 1 

cnt=$(ps -ef | grep chatroom | grep -v grep | wc -l)
if [ ${cnt} -gt 1 ];then 
    echo "start chatroom success."
else 
    echo "start maybe failed."
fi

