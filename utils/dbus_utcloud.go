package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

const (
	utcloudDBusService = "com.deepin.utcloud.Daemon"
	utcloudDBusPath    = "/com/deepin/utcloud/Daemon"
	utcloudServer      = "http://utcloud-pre.chinauos.com"
)

type UtResponse struct {
	Code int `json:"code"`
	Data struct {
		Id string `json:"id"`
	} `json:"data"`
	Err struct {
	} `json:"err"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}

type OsInfo struct {
	OsName    string `json:"os_name" structs:"os_name"`
	OsEdition string `json:"os_edition" structs:"os_edition"`
	OsVerion  string `json:"os_version" structs:"os_version"` //MajorVersion.MinorVersion
}

// AddOS
//  @Description: 添加系统
//  @param token string
//  @return bool
//  @return string
//
func AddOS(token string) (string, error) {
	head := map[string]string{
		"token": token,
	}
	viper.SetConfigName("os-version")
	viper.AddConfigPath("/etc/")
	viper.SetConfigType("ini")
	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}
	params := map[string]interface{}{
		"os_edition": viper.GetString("Version.EditionName"),
		"os_name":    "uos",
		"os_version": fmt.Sprintf("%s.%s", viper.GetString("Version.MajorVersion"), viper.GetString("Version.MinorVersion")),
	}
	data, err := HTTPCall(http.MethodPut, utcloudServer+"/api/v0/access/os", head, nil, params)
	if err != nil {
		log.Errorf("add os error:[%s]", err.Error())
		return "", err
	}
	var response = UtResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("add os response error:[%s]", err.Error())
		return "", err
	}
	log.Debugf("add os response :[%#v]", response)
	if !response.Result {
		return "", errors.New(response.Msg)
	}
	return response.Data.Id, nil
}

// AddApp 添加应用
//  @Description: 添加应用
//  @param token string
//  @return bool
//  @return string
//
func AddApp(token string) (string, error) {
	head := map[string]string{
		"token": token,
	}
	binPath, _ := GetRunPath()
	params := map[string]interface{}{
		"callback_dbus_method": ServiceName + ".Callback",
		"callback_dbus_name":   ServiceName,
		"callback_dbus_path":   ServicePath,
		"description":          "测试用demo",
		"developer":            "ut003500",
		"email":                "ut003500@uniontech.com",
		"name":                 "测试云服务对接app",
		"path":                 binPath,
		"show_switcher":        false,
	}
	data, err := HTTPCall(http.MethodPut, utcloudServer+"/api/v0/access/app", head, nil, params)
	if err != nil {
		log.Errorf("add os error:[%s]", err.Error())
		return "", err
	}
	var response = UtResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("add app host response error:[%s]", err.Error())
		return "", err
	}
	log.Debugf("add app response :[%#v]", response)
	if !response.Result {
		return "", errors.New(response.Msg)
	}
	return response.Data.Id, nil
}

// BindApp2OS
//  @Description: 绑定app到系统
//  @param token string 云服务token
//  @param osId string 已绑定osId
//  @param appId string 已添加的appId
//  @return bool
//  @return error
//
func BindApp2OS(token, osId, appId string) (bool, error) {
	head := map[string]string{
		"token": token,
	}
	params := map[string]interface{}{
		"appids": []string{
			appId,
		},
		"osid": osId,
	}
	data, err := HTTPCall(http.MethodPut, utcloudServer+"/api/v0/access/attach", head, nil, params)
	if err != nil {
		log.Errorf("bind app to os error:[%s]", err.Error())
		return false, err
	}
	var response = UtResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("bind app to os host response error:[%s]", err.Error())
		return false, err
	}
	log.Debugf("bind app to os response [%#v]", response)
	if !response.Result {
		return false, errors.New(response.Msg)
	}
	return true, nil
}

type uploadResponse struct {
	Code int `json:"code"`
	Data struct {
		Acl struct {
			SignUrl string `json:"sign_url"`
		} `json:"acl"`
		Uss int `json:"uss"`
	} `json:"data"`
	Err struct {
	} `json:"err"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}

//
//  uploadFile
//  @Description:通过临时授权上传文件
//  @param token
//  @param fName
//
func uploadFile(token, fName string) (bool, error) {
	head := map[string]string{
		"token": token,
	}
	_, hash, _ := md5sumAndsize(fName, true)
	binPath, _ := GetRunPath()
	params := map[string]interface{}{
		"bin_path": binPath,
		"key":      fName,
		"method":   "put",
		"md5":      hash,
	}
	data, err := HTTPCall(http.MethodGet, utcloudServer+"/api/v0/app/acl", head, params, nil)
	if err != nil {
		log.Errorf("upload file error:[%s]", err.Error())
		return false, err
	}
	var response = uploadResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("upload file host response error:[%s]", err.Error())
		return false, err
	}
	log.Debugf("upload file response [%#v]", response)
	if !response.Result {
		return false, errors.New(response.Msg)
	}
	bucket, err := GetBucket(response.Data.Acl.SignUrl)
	if err != nil {
		return false, err
	}

	err = PutObject(bucket, response.Data.Acl.SignUrl, fName, hash)
	if err != nil {
		return false, err
	}
	return true, nil
}

//
//  noteMetaData
//  @Description: 上传成功后通知服务端
//  @param token
//  @param fName
//  @return []byte
//  @return error
//
func noteMetaData(token, fName string) ([]byte, error) {
	head := map[string]string{
		"token": token,
	}
	binPath, _ := GetRunPath()
	params := map[string]interface{}{
		"bin_path": binPath,
		"key":      fName,
	}
	data, err := HTTPCall(http.MethodPut, utcloudServer+"/api/v0/app/meta", head, nil, params)
	if err != nil {
		log.Errorf("note upload error:[%s]", err.Error())
		return []byte(""), err
	}
	var response = UtResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("note upload host response error:[%s]", err.Error())
		return []byte(""), err
	}
	log.Debugf("note upload response [%#v]", response)
	if !response.Result {
		return []byte(""), errors.New(response.Msg)
	}
	return []byte(response.Data.Id), nil
}

//
// UploadUtDaemon
//  @Description: 上传utcloud服务文件
//  @param fName
//  @return bool
//  @return error
//
func UploadUtDaemon(token, fName string) ([]byte, error) {
	ok, err := uploadFile(token, fName)
	if err != nil {
		return []byte(""), err
	}
	if !ok {
		return []byte(""), errors.New("upload fail without error")
	}
	return noteMetaData(token, fName)
}

//
// UploadByDaemon
//  @Description: 通过utcloud上传文件
//  @param fName
//  @return bool
//  @return error
//
func UploadByDaemon(fName string) ([]byte, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warnf("init conn error.%v", err)
		return []byte(""), errors.New("init conn error")
	}
	var s []byte
	object := conn.Object("com.deepin.utcloud.Daemon", "/com/deepin/utcloud/Daemon")
	err = object.Call("Upload", 0, fName).Store(&s)
	if err != nil {
		fmt.Println("upload file by utcloud daemon fail:", err)
		return []byte(""), err
	}
	return s, nil
}

type deleteResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Err  struct {
	} `json:"err"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}

//
// DeleteDaemon
//  @Description: 删除daemon文件
//  @param fName
//  @return bool
//  @return error
//
func DeleteDaemon(token, fName string) (bool, error) {
	head := map[string]string{
		"token": token,
	}
	binPath, _ := GetRunPath()
	params := map[string]interface{}{
		"bin_path": binPath,
		"key":      fName,
	}
	data, err := HTTPCall(http.MethodDelete, utcloudServer+"/api/v0/app/meta", head, nil, params)
	if err != nil {
		log.Errorf("delete file error:[%s]", err.Error())
		return false, err
	}
	var response = deleteResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		log.Errorf("delete file host response error:[%s]", err.Error())
		return false, err
	}
	log.Debugf("delete file response [%#v]", response)
	if !response.Result {
		return false, errors.New(response.Msg)
	}
	return response.Result, nil
}

//
// DeleteByDaemon
//  @Description: 删除daemon文件
//  @param fName
//  @return bool
//  @return error
//
func DeleteByDaemon(fName string) (string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warnf("init conn error.%v", err)
		return "", errors.New("init conn error")
	}
	var s string
	object := conn.Object("com.deepin.utcloud.Daemon", "/com/deepin/utcloud/Daemon")
	err = object.Call("Delete", 0, fName).Store(&s)
	if err != nil {
		fmt.Println("delete file by utcloud daemon fail:", err)
		return "", err
	}
	return s, nil
}
