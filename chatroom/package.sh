#!/bin/bash

project_name="chatroom"

# build
sh build.sh

# package
version=`./chatroom -version | grep Version | awk '{print $2}'`
distname=${project_name}_${version}

rm -rf package
mkdir -p package/${project_name}
cp ${project_name} package/${project_name}/
cp config.ini package/${project_name}/
cp -rf html package/${project_name}/
cp start.sh package/${project_name}/
cp stop.sh package/${project_name}/
cp status.sh package/${project_name}/
cp README.md package/${project_name}/
cd package

tar -zcvf $distname.tar.gz ${project_name} \
      --exclude .svn --exclude .git --exclude *.log* 
cd ..
rm -rf package/${project_name}/
