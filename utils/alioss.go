package utils

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func GetBucket(aliyunSign string) (*oss.Bucket, error) {
	client, err := oss.New(aliyunSign, "", "")
	if err != nil {
		log.Errorf("oss create empty client error:[%s]", err.Error())
		return nil, errors.New("oss opt error")
	}

	client.Config.IsEnableCRC = false //取消crc32校验
	client.Config.IsEnableMD5 = true  //使用md5校验
	bucket := &oss.Bucket{Client: *client}
	return bucket, nil
}

//
// GetObject
//  @Description: 下载数据:
//  @param bucket
//  @param signUrl
//  @param localFile
//  @return error
//
func GetObject(bucket *oss.Bucket, signUrl, localFile string) error {
	if bucket == nil {
		return errors.New("param invalid")
	}

	if err := CheckWriteFile(localFile); err != nil {
		log.Errorf("check local path:[%s] error", localFile)
		return err
	}

	err := bucket.GetObjectToFileWithURL(signUrl, localFile)
	if err != nil {
		serr, ok := err.(oss.ServiceError)
		if !ok {
			log.Errorf("get object error:[%s]", err.Error())
			if strings.Contains(err.Error(), "no such file or directory") {
				return errors.New("no such file or directory")
			}
			return err
		}
		if serr.StatusCode == http.StatusNotFound {
			err = errors.New("record not exist")
		}
	}
	return err
}

//
// PutObject
//  @Description: 上传数据:
//  @param bucket
//  @param signUrl
//  @param localFile
//  @param md5sum
//  @return error
//
func PutObject(bucket *oss.Bucket, signUrl, localFile, md5sum string) error {
	if bucket == nil {
		return errors.New("param invalid")
	}

	bmd5, _ := hex.DecodeString(md5sum)
	opts := []oss.Option{
		oss.ContentMD5(base64.StdEncoding.EncodeToString(bmd5)),
	}
	// opts := oss.AddContentType(nil, localFile)
	// opts = append(opts, oss.ContentMD5(base64.StdEncoding.EncodeToString(bmd5)))

	err := bucket.PutObjectFromFileWithURL(signUrl, localFile, opts...)
	if err == nil {
		return err
	}

	log.Errorf("alioss put object error:[%s]", err.Error())
	if strings.Contains(err.Error(), "no such file or directory") {
		return errors.New("no such file or directory")
	}
	return errors.New("oss opt error")
}

//
// DeleteObjects
//  @Description: 删除数据
//  @param bucket
//  @param paths
//  @return []string
//  @return error
//
func DeleteObjects(bucket *oss.Bucket, paths []string) ([]string, error) {
	// 返回删除成功的文件。
	delRes, err := bucket.DeleteObjects(paths)
	if err != nil {
		if err.(oss.ServiceError).StatusCode == http.StatusNotFound {
			err = errors.New("record not exist")
		}
	}
	return delRes.DeletedObjects, err
}
