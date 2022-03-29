package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	TagCookie string = "token"
)

type _errSt struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e _errSt) Error() string {
	be, _ := json.Marshal(&e)
	return string(be)
}

// newError
func newError(r int, m string) error {
	return _errSt{Code: r, Msg: m}
}

// HTTPCall curl请求
func HTTPCall(method, remoteURL string, head map[string]string, query, body map[string]interface{}) ([]byte, error) {
	log.Debugf("http request method is:[%s], remoteURL is:[%s], head is:[%#v] query is:[%s] body is:[%#v]", method, remoteURL, head, query, body)
	var err error = nil
	var bd []byte
	if body != nil {
		bd, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, remoteURL, bytes.NewBuffer(bd))
	if err != nil {
		return nil, err
	}

	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, fmt.Sprint(v))
		}
		req.URL.RawQuery = q.Encode()
	}
	for k, v := range head {
		if strings.EqualFold(k, TagCookie) {
			req.AddCookie(&http.Cookie{Name: k, Value: v, HttpOnly: true})
		}
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("http request do error:[%s]", err.Error())
		return nil, newError(resp.StatusCode, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, newError(resp.StatusCode, "response status code error")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newError(resp.StatusCode, err.Error())
	}

	return data, nil
}
