package utils

import (
	"errors"
	"github.com/godbus/dbus/v5"
	"github.com/jandre/procfs"
	log "github.com/sirupsen/logrus"
)

const orgFreedesktopDBus = "org.freedesktop.DBus"

func GetDbusSender(name string) (string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warnf("init conn error.%v", err)
		return "", errors.New("init conn error")
	}

	pid, err := GetConnPID(conn, name)
	if err != nil {
		log.Errorf("CheckSender get pid error:[%s]", err.Error())
		return "", err
	}

	path, err := GetProcessPath(int(pid))
	if err != nil {
		log.Errorf("CheckSender get exe path error:[%s]", err.Error())
	}
	return path, err
}

// GetConnPID 获取conn进程ID
func GetConnPID(conn *dbus.Conn, name string) (pid uint32, err error) {
	err = conn.BusObject().Call(orgFreedesktopDBus+".GetConnectionUnixProcessID",
		0, name).Store(&pid)
	return
}

// GetProcessPath 获取conn进程path
func GetProcessPath(pid int) (exe string, err error) {
	process, err := procfs.NewProcess(pid, true)
	return process.Exe, err
}
