package service

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/harveywangdao/ants/logger"
	"github.com/harveywangdao/ants/util"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	APPID     = "wx374a0f37b50eb5ce"
	APPSECRET = "1ffa637e5224634a85228c825289c548"
)

type WxUserInfo struct {
	Openid     string `json:"openid,omitempty"`
	SessionKey string `json:"session_key,omitempty"`
	Unionid    string `json:"unionid,omitempty"`
	Errcode    int    `json:"errcode,omitempty"`
	Errmsg     string `json:"errmsg,omitempty"`
}

func code2Session(code string) (*WxUserInfo, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", APPID, APPSECRET, code)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(string(body))

	wxUserInfo := WxUserInfo{}
	if err := json.Unmarshal(body, &wxUserInfo); err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info("wxUserInfo", wxUserInfo)

	if wxUserInfo.Errcode != 0 {
		logger.Error("code2Session fail, Errmsg:", wxUserInfo.Errmsg)
		return nil, fmt.Errorf("code2Session fail, Errmsg: " + wxUserInfo.Errmsg)
	}

	return &wxUserInfo, nil
}

func checkWxUserInfoSign(rawData, sessionKey, sign string) bool {
	sh1 := sha1.Sum([]byte(rawData + sessionKey))
	sign2 := hex.EncodeToString(sh1[:])
	sign = strings.ToLower(sign)
	sign2 = strings.ToLower(sign2)
	return sign == sign2
}

func decryptWxUserInfo(encryptedData, iv, sessionKey string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	plaintext, err := util.AesCBCDecrypt(ciphertext, key, ivBytes)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return plaintext, nil
}
