#!/usr/bin/env bash

wget https://oss.iinti.cn/majora/bin/latest/majora-cli_latest_linux_amd64.tar.gz -O majora-cli.tar.gz

rm -fr majora
rm -fr exec.sh
rm -fr start.sh
rm -fr majora.yaml
rm -fr majora.service
rm -fr majora-dev.yaml
rm -fr majora.log
rm -fr majora-log
rm -fr std.log

tar -zxvf majora-cli.tar.gz

mv -f majora-cli*/* .

exec bash ./start.sh