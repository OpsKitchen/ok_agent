#!/bin/sh
#
# ok_agent - a linux agent of OpsKitchen.com
#
# chkconfig:   - 58 42
# description: a linux agent of OpsKitchen.com
# processname:  ok_agent
# config:       /etc/ok_agent.json
#

### BEGIN INIT INFO
# Provides: OpsKitchen.com
# Required-Start: $local_fs $network $remote_fs
# Required-Stop: $local_fs $network $remote_fs
# Default-Stop: 0 1 2 6
# Short-Description: start and stop ok_agent
# Description: a linux agent of OpsKitchen.com
### END INIT INFO

# Source function library.
. /etc/rc.d/init.d/functions

# Source networking configuration.
. /etc/sysconfig/network

# Check that networking is up.
[ "$NETWORKING" = "no" ] && exit
exec="/usr/sbin/ok_agent"
prog=$(basename $exec)

lockfile=/var/lock/subsys/ok_agent

start() {
    [ -x $exec ] || exit 5
    echo -n $"Starting $prog: "
    # if not running, start it up here, usually something like "daemon $exec"

    daemon /usr/sbin/daemonize $exec
    retval=$?
    echo
    [ $retval -eq 0 ] && touch $lockfile
    return $retval
}

stop() {
    echo -n $"Stopping $prog: "
    # stop it here, often "killproc $prog"
    killproc $prog -INT
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    start
}

reload() {
    restart
}

force_reload() {
    restart
}

rh_status() {
    # run checks to determine if the service is running or use generic status
    status $prog
}

rh_status_q() {
    rh_status >/dev/null 2>&1
}


case "$1" in
    start)
        rh_status_q && exit 0
        $1
        ;;
    stop)
        rh_status_q || exit 0
        $1
        ;;
    restart)
        $1
        ;;
    reload)
        rh_status_q || exit 7
        $1
        ;;
    force-reload)
        force_reload
        ;;
    status)
        rh_status
        ;;
    condrestart|try-restart)
        rh_status_q || exit 0
        restart
        ;;
    *)
        echo $"Usage: $0 {start|stop|status|restart|condrestart|try-restart|reload|force-reload}"
        exit 2
esac
exit $?