#!/usr/bin/env bash
if [ ! -f `basename $0` ]; then
    echo "please run me in my container folder"
    exit 1
fi

if [ -z $GOPATH ]; then
    echo "GOPATH not set"
    exit 1
fi

prod_name="ok_agent"
src_dir="/root/rpmbuild/SOURCES"
cwd=`pwd`
version=`grep AgentVersion ok_agent.json | awk -F '"' '{print $4}'`
prod_name_with_version=${prod_name}-${version}


#prepare src dir
mkdir -p ${src_dir}

#compress source
cd ${src_dir}
mkdir -p ${prod_name_with_version}
cp -r ${cwd}/* ${prod_name_with_version}
tar zcf ${prod_name_with_version}.tar.gz ${prod_name_with_version}/
rm -rf ${prod_name_with_version}/*
cd ${cwd}

echo "
Summary:    OpsKitchen.com linux agent rpm package
Name:       ${prod_name}
Version:    ${version}
Release:    1
Source:     ${prod_name_with_version}.tar.gz
License:    GPL
Packager:   qinjx
Group:      Application
URL:        http://www.OpsKitchen.com

%description
This is the linux agent for OpsKitchen.com

%prep
%setup -q

%build
go build

%install
rm -rf \$RPM_BUILD_ROOT
mkdir -p \$RPM_BUILD_ROOT/usr/sbin
mkdir -p \$RPM_BUILD_ROOT/etc
mkdir -p  \$RPM_BUILD_ROOT/root/.ok_agent

install -m 755 ok_agent \$RPM_BUILD_ROOT/usr/sbin/ok_agent
install -m 755 ok_agent.json \$RPM_BUILD_ROOT/etc/ok_agent.json
install -m 400 credential.json \$RPM_BUILD_ROOT/root/.ok_agent/credential.json

%files
/usr/sbin/ok_agent
/etc/ok_agent.json

%config /root/.ok_agent/credential.json

" > ${cwd}/${prod_name}-${version}.spec

#build rpm
rpmbuild -bb ${cwd}/${prod_name}-${version}.spec