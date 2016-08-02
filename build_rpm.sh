#!/usr/bin/env bash
if [ ! -f `basename $0` ]; then
    echo "please run me in my container folder"
    exit 1
fi

if [ -z $GOPATH ]; then
    echo "GOPATH not set"
    exit 1
fi

cwd=`pwd`
prod_name=`basename ${cwd}`
src_dir="/root/rpmbuild/SOURCES"
version=`grep AgentVersion ./model/config/config.go | awk -F '"' '{print $2}'`
prod_name_with_version=${prod_name}-${version}
src_tar_gz=${prod_name_with_version}.tar.gz
spec_file=${src_dir}/${prod_name}-${version}.spec

#compile go code
go build

#prepare src dir
mkdir -p ${src_dir}

#compress source
cd ${src_dir}
rm -rf ${prod_name_with_version}/
mkdir -p ${prod_name_with_version}
cp -r ${cwd}/* ${prod_name_with_version}/
tar zcf ${src_tar_gz} ${prod_name_with_version}/
cd ${cwd}

echo "
Summary:    OpsKitchen.com linux agent rpm package
Name:       ${prod_name}
Version:    ${version}
Release:    1
Source:     ${src_tar_gz}
License:    GPL
Packager:   qinjx
Group:      Application
URL:        http://www.OpsKitchen.com

%description
This is the linux agent for OpsKitchen.com

%prep
%setup -q

%build
#go build

%install
rm -rf \$RPM_BUILD_ROOT
mkdir -p \$RPM_BUILD_ROOT/usr/sbin
mkdir -p \$RPM_BUILD_ROOT/etc/init.d
mkdir -p \$RPM_BUILD_ROOT/root/.ok_agent

install -m 755 ok_agent        \$RPM_BUILD_ROOT/usr/sbin/ok_agent
install -m 755 init.d.sh       \$RPM_BUILD_ROOT/etc/init.d/ok_agent
install -m 400 credential.json \$RPM_BUILD_ROOT/root/.ok_agent/credential.json

%files
/usr/sbin/ok_agent
/etc/init.d/ok_agent

%config /root/.ok_agent/credential.json

" > ${spec_file}

#build rpm
rpmbuild -bb ${spec_file}