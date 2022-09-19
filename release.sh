#!/usr/bin/env bash

make
if [ $? -ne 0 ]; then
    echo "make error"
    exit 1
fi

majora_version=`./majora version`
echo "build version: $majora_version"

# cross_compiles
make -f ./Makefile.cross

rm -rf ./release/packages
mkdir -p ./release/packages

os_all='linux windows darwin freebsd'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle'

cd ./release

for os in $os_all; do
    for arch in $arch_all; do
        majora_dir_name="majora-cli_${majora_version}_${os}_${arch}"
        majora_path="./packages/majora-cli_${majora_version}_${os}_${arch}"

        if [ "x${os}" = x"windows" ]; then
            if [ ! -f "./majora_${os}_${arch}.exe" ]; then
                continue
            fi
            mkdir ${majora_path}
            mv ./majora_${os}_${arch}.exe ${majora_path}/majora.exe
        else
            if [ ! -f "./majora_${os}_${arch}" ]; then
                continue
            fi
            mkdir ${majora_path}
            mv ./majora_${os}_${arch} ${majora_path}/majora
        fi
        cp -rf ../conf/* ${majora_path}

        # packages
        cd ./packages
        if [ "x${os}" = x"windows" ]; then
            zip -rq ${majora_dir_name}.zip ${majora_dir_name}
            scp "${majora_dir_name}.zip" root@oss.iinti.cn:/root/gohttpserver/data/majora/bin/"$majora_version"
        else
            tar -zcf ${majora_dir_name}.tar.gz ${majora_dir_name}
            scp "${majora_dir_name}.tar.gz" root@oss.iinti.cn:/root/gohttpserver/data/majora/bin/"$majora_version"
        fi
        cd ..
        rm -rf ${majora_path}
    done
done

cd -
