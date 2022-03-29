package service

import (
	"errors"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	log "github.com/sirupsen/logrus"
	"utMsgDaemon/utils"
)

var stopped = make(chan struct{}, 1)

func initDbusService() {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalf("connect session bus fail: %v", err)
		return
	}
	reply, err := conn.RequestName(utils.ServiceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Fatalf("query service name state fail:%v", err)
		return
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalln("service name already exists!")
		return
	}
	s := &Service{
		ID:    "2",
		Name:  "lisi",
		Index: 1,
	}

	err = conn.Export(s, utils.ServicePath, utils.ServiceInterface)
	if err != nil {
		log.Fatalf("export service fail:%v", err)
		return
	}

	props, err := utils.NewProperty(s)
	if err != nil {
		log.Fatalf("export prop fail:%v", err)
		return
	}

	mp, _ := utils.NewMulti()
	p, _ := mp.Add(utils.ServiceInterface, s)

	node := introspect.Node{
		Name: utils.ServicePath,
		Interfaces: []introspect.Interface{
			{
				Name:       utils.ServiceInterface,
				Methods:    introspect.Methods(s),
				Properties: p.Introspection(),
			},
			props.Interface(),
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&node), utils.ServicePath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Fatalf("export service info to dbus fail:%v", err)
		return
	}
	err = mp.Export(conn, utils.ServicePath)
	if err != nil {
		log.Fatalf("export service props to dbus fail:%v", err)
		return
	}
	select {
	case <-stopped:
		return
	}
}

type Service struct {
	ID    string `dbus:"const,emit"`
	Name  string `dbus:"writeable,emit"`
	Token string `dbus:"writeable,emit"`
	Index int    `dbus:"writeable,emit"`
}

func (s *Service) SetToken(value string) *dbus.Error {
	s.Token = value
	return nil
}

func (s *Service) Callback(sender dbus.Sender, mType int16, fName string, fSign string) (bool, *dbus.Error) {
	path, err := utils.GetDbusSender(string(sender))
	if err != nil {
		return false, utils.NewError(err).Error
	}
	log.WithFields(log.Fields{
		"path":   path,
		"method": "callback",
	})
	log.Debugf("回调接口推送信息为： %d %s %s", mType, fName, fSign)
	return true, nil
}

// AddWhitelist 给当前机器添加白名单
func (s *Service) AddWhitelist() (bool, *dbus.Error) {
	if len(s.Token) == 0 {
		log.Warn("not token found")
		return false, utils.NewError(errors.New("token not found")).Error
	}
	osId, err := utils.AddOS(s.Token)
	if err != nil {
		return false, utils.NewError(err).Error
	}
	appId, err := utils.AddApp(s.Token)
	if err != nil {
		return false, utils.NewError(err).Error
	}
	ok, err := utils.BindApp2OS(s.Token, osId, appId)
	if err != nil {
		return false, utils.NewError(err).Error
	}
	return ok, nil
}

//Upload 云服务上传文件
func (s *Service) Upload(sender dbus.Sender, key string) ([]byte, *dbus.Error) {
	path, err := utils.GetDbusSender(string(sender))
	if err != nil {
		return []byte(""), utils.NewError(err).Error
	}
	bts, err := utils.UploadByDaemon(key)
	if err != nil {
		log.WithFields(log.Fields{
			"path":   path,
			"method": "Upload",
		})
		log.Errorf("上传出错，错误信息为：%#v %#v", bts, err)
		return []byte(""), utils.NewError(err).Error
	}
	return bts, nil
}

//Delete 云服务删除文件
func (s *Service) Delete(sender dbus.Sender, key string) (string, *dbus.Error) {
	path, err := utils.GetDbusSender(string(sender))
	if err != nil {
		return "", utils.NewError(err).Error
	}
	str, err := utils.DeleteByDaemon(key)
	if err != nil {
		log.WithFields(log.Fields{
			"path":   path,
			"method": "Delete",
		})
		log.Errorf("删除出错，错误信息为: %#v", err)
		return "", utils.NewError(err).Error
	}
	return str, nil
}
