#!/bin/bash

configPath=$1

slaveInfo=`grep -Po 'ID[" :]+\K[^"]+' $configPath | grep S`
slaveInfoArr=(${slaveInfo})
# echo ${slaveInfoArr[@]}

slaveIPInfo=`grep -Po 'HOST[" :]+\K[^"]+' $configPath | grep 172`
slaveIPInfoArr=(${slaveIPInfo})
# echo ${slaveIPInfoArr[@]}

localIPInfo=`ifconfig | grep inet | grep -v inet6 | grep -v 127 | cut -d ' ' -f10`
# echo $localIPInfo

# start slave
for index in ${!slaveInfoArr[@]}; do
# echo $index
if [ ${slaveIPInfoArr[index]} == $localIPInfo ]; then
	 cmd="./cluster --cluster_config "${configPath}" --service "${slaveInfoArr[index]}">> "${slaveInfoArr[index]}".log 2>&1 &"
	 echo $cmd
	 eval $cmd
fi
done


sleep 10s

# start master
cmd="./cluster --cluster_config "${configPath}" >>master.log 2>&1 &"
echo $cmd
eval $cmd
