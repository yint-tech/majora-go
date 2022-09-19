#!/usr/bin/env bash

old_pid=$(ps -ef | grep majora | grep -v grep | awk '{print $2}')

echo "old pid is ${old_pid}"

echo "clean old ..."

`ps -ef | grep majora | grep -v grep | awk '{print $2}'| xargs kill -9`
mkdir -p "majora-log"

exec ./majora -conf majora.yaml