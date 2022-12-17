#!/bin/sh

old_pid=$(ps | grep majora |grep -v grep | awk '{print $1}')

echo "old pid is ${old_pid}"

echo "clean old ..."

`ps | grep majora |grep -v grep | awk '{print $1}'  | xargs kill -9`
mkdir -p "majora-log"

exec ./majora -conf majora.yaml
