$GOARCH    目标平台（编译后的目标平台）的处理器架构（386、amd64、arm）
$GOOS          目标平台（编译后的目标平台）的操作系统（darwin、freebsd、linux、windows）

各平台的GOOS和GOARCH参考 

OS                   ARCH                          OS version
linux                386 / amd64 / arm             >= Linux 2.6
darwin               386 / amd64                   OS X (Snow Leopard + Lion)
freebsd              386 / amd64                   >= FreeBSD 7
windows              386 / amd64                   >= Windows 2000


cmd:
GOOS=linux GOARCH=arm go build -x -o client_linux_arm client.go
GOOS=linux GOARCH=amd64 go build -x -o client_linux_amd64 client.go
GOOS=linux GOARCH=amd64 go build -x -o agent client.go
