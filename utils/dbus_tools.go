package utils

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const (
	ServiceName      = "com.uniontech.msgExample"
	ServicePath      = "/com/uniontech/msgExample"
	ServiceInterface = "com.uniontech.msgExample"
)

// UserInfo convert dbus UserInfo prop
type UserInfo struct {
	Uid          string
	Username     string
	Nickname     string
	ProfileImage string
	Region       string
	HardwareID   string
	IsLoggedIn   bool
}

// FromDBus convert map variant to struct
func (u *UserInfo) FromDBus(v interface{}) {
	vMap, _ := v.(dbus.Variant)
	dest := map[string]interface{}{}
	_ = dbus.Store([]interface{}{vMap}, &dest)
	_ = mapstructure.Decode(dest, &u)
}

// IsValid check user info is valid
func (u *UserInfo) IsValid() bool {
	uid, _ := strconv.Atoi(u.Uid)
	return uid > 0
}

type WrapError struct {
	*dbus.Error
}

func NewError(err error) WrapError {
	if nil == err {
		return WrapError{
			Error: nil,
		}
	}
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	funcFullName := details.Name()
	names := strings.Split(funcFullName, ".")
	method := ServiceName + "." + names[len(names)-1]
	return WrapError{
		Error: dbus.NewError(method, []interface{}{err.Error()}),
	}
}

const bufferSize = 1024 * 4

// returns MD5 checksum of filename and it"s size
func md5sumAndsize(filename string, m5 bool) (int64, string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		log.Errorf("md5sumAndsize stat error:[%#v]", err)
		return 0, "", errors.New("file opt error")
	} else if info.IsDir() {
		return 0, "", errors.New("path or keys should be file")
	}

	if !m5 {
		return info.Size(), "", nil
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("md5sumAndsize open error:[%#v]", err)
		return 0, "", errors.New("file opt error")
	}
	defer file.Close()

	hash := md5.New()
	for buf, reader := make([]byte, bufferSize), bufio.NewReader(file); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, "", errors.New("file opt error")
		}

		hash.Write(buf[:n])
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return info.Size(), checksum, nil
}

func CheckWriteFile(v string) error {
	if len(v) == 0 {
		panic("xxxxx")
	}

	var isDir, canWrite bool
	err := FullPathCheck(v, &isDir, &canWrite)
	if err == nil {
		if canWrite {
			return nil
		}
		if isDir {
			return errors.New("path or keys should be file")
		}
		return errors.New("file opt error")
	}

	if !os.IsNotExist(err) {
		log.Errorf("CheckWriteFile stat error:[%#v]", err)
		return errors.New("file opt error")
	}

	// 不存在，创建
	i := strings.LastIndex(v, "/")
	dir := v[:i]
	err = MakeDir(dir)
	if err != nil {
		return errors.New("file opt error")
	}
	return nil
}

func MakeDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Errorf("path mkdir error:[%s] ", err.Error())
			return err
		}
		return nil
	}
	return err
}

func FullPathCheck(v string, isdir, canwrite *bool) error {
	if len(v) == 0 {
		panic("xxxxx")
	}

	info, err := os.Stat(v)
	if err != nil {
		return err
	}

	*isdir = false
	*canwrite = false

	im := info.Mode()
	if im.IsDir() { // directory ?
		*isdir = true
	}

	pu, _ := user.Current()
	sysd := info.Sys().(*syscall.Stat_t)
	imp := im.Perm()

	// all write,  --------w- : 00 0000 0010 =  0x02
	if imp&0x02 == 0x02 {
		fmt.Println("all user can write")
		*canwrite = true
		return nil
	}

	// gid write,  -----w---- : 00 0001 0000 =  0x10
	if imp&0x10 == 0x10 && pu.Gid == fmt.Sprintf("%d", sysd.Gid) {
		fmt.Println("same group can write")
		*canwrite = true
		return nil
	}

	// uid write,  --w------- : 00 1000 0000 =  0x80
	if imp&0x80 == 0x80 && pu.Uid == fmt.Sprintf("%d", sysd.Uid) {
		fmt.Println("same uid can write")
		*canwrite = true
		return nil
	}

	return nil
}

func GetRunPath() (string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return path, err
}
