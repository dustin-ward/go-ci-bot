#!/bin/bash

# Check to see if the bot is already running
declare -a aa
me=$(/bin/basename $0)
while read line; do
        aa=($line)
        cmd=$(/bin/basename ${aa[3]})
        if [[ ${cmd} == ${me} ]]; then
                if [[ ${aa[0]} != $$ ]]; then
                        if [[ ${aa[1]} != $$ ]]; then
                                echo ${aa[0]} ${aa[1]} $me is already running
                                exit
                        fi
                fi
        fi
done < <(/bin/ps -C $me -o pid=,ppid=,args=)

# Pull any new changes and re-build the bot
cd /home/zosgo/jeff-ci/ || exit 1
git pull
go build

# Run the bot with nohup and send logs to nohup.out
nohup ./jeff-ci
