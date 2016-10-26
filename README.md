# ok_gent
Linux agent for OpsKitchen.com

## Get started
### Install requirements
    yum install -y libxml2-devel augeas-devel gcc git

### Prepare Go SDK
Download go SDK 1.7 from: https://golang.org/dl/

Decompress it into /usr/local/go

Run:

    export PATH=$PATH:/usr/local/go/bin
    export GOROOT="/usr/local/go/"

#### Go version
Make sure to download the 1.7 version (beta is also ok).

### Prepare working dir

    mkdir /opt/go
    cd /opt/go
    export GOPATH=`pwd`


### Download this project:

    cd $GOPATH
    go get github.com/OpsKitchen/ok_agent

### Build binary

    cd src/github.com/OpsKitchen/ok_agent
    go build

### Customize your demo data

Edit ok_agent.json, credential.json, replace the demo data

### Run

    ./ok_agent -c ok_agent.json

Demo output is like this:

    [root@dev67 ok_agent]# ./ok_agent -c ok_agent.json
    INFO[0000] Runing opskitchen agent 1.1.0
    INFO[0000] Succeed to call entrance api.
    INFO[0000] Processing file: /etc/yum.repos.d/test.repo
    INFO[0000] Processing file: /etc/php.ini
    INFO[0000] Processing file: /etc/sysconfig/beanstalkd
    INFO[0000] Processing Augeas: PHP/display_errors@/etc/php.ini
    INFO[0000] Processing Augeas: BEANSTALKD_BINLOG_DIR@/etc/sysconfig/beanstalkd
    INFO[0000] Processing file: /tmp/agent_test/dir_mysql
    INFO[0000] Processing file: /tmp/agent_test/file with space
    INFO[0000] Processing file: /tmp/agent_test/file_empty
    INFO[0000] Processing file: /tmp/agent_test/no_such_dir/file_with_content
    INFO[0000] Processing file: /tmp/agent_test/link_test
    INFO[0000] Processing command: gofmt -w /opt/ok/ok_agent
    INFO[0000] Running command...
    INFO[0000] Succeed to run command.
    INFO[0000] Processing command: touch mysql.txt
    INFO[0000] Running command...
    INFO[0000] Succeed to run command.
    INFO[0000] Congratulations! All tasks have been done successfully!