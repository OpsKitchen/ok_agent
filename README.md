# ok_agent
ok_agent is an agent of OpsKitchen.com, running on customers' OS.

We release the source code for community review.

If you find any bug, or potential security problem, please create an issue to let us know.

1、main dir : src/client
2、config file : src/config/baseConfig.go
3、Folder Structure
agent
	|--bin
	|--pkg
	|--src  
		|--client
			|--client.go (main)

		|--config
			|--baseConfig.go

		|--okAgent
			|--augeas
			|--command
			|--file
			|--http

		|--honnef.co
			|--augeas …

		|--github.com
			|--bitly …
			|--pantsing …

		|--logger
			|--logger.go 日志

-



src
    |--main.go (命令行交互)
    |--Agent.go (主class)
    |--Adapter
        |--File.go
        |--Command.go
        |--ConfigOption
            |--Augeas.go
    |--Util
        |--Log (日志)
        |--Term (终端下输出彩色)
    |--vendor
        |--honnef.co
            |--go
                |--augeas
        |--github.com (自动下载)
            |--OpsKitchen
                |--ok_api_sdk
                    |--go
asset
    |--ok_agent.conf (credential文件地址,base_api名称,网关地址)
    |--ok_agent.credential (secret, server_unique_name)
    |--message
        |--en_US
            |--main.json
            |--api_result.json (API错误码相关错误信息)
script
    |--ok_agent.spec (rpm包定义)
    |--build_rpm.sh (打包脚本)
    |--install_deps.sh (安装配置编译环境)
    |--Make.m4