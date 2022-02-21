package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	baseUrl := "http://127.0.0.1:8080/api/open"
	appKey := "appKey"
	appSecret := "appSecret"

	sdk := NewSdk(baseUrl, appKey, appSecret)
	_ = sdk.GetDevices(67, "20210901103050484")
	_ = sdk.SaveTestData("20210901103050484", []map[string]interface{}{
		{
			"test_item_id": 10000,
			"test_data": []map[string]interface{}{
				{
					"item1": 1,
					"item2": "2",
					"item3": true,
				},
			},
		},
	})
	_ = sdk.UploadImage("baidu.png")
}

type Sdk struct {
	client    *http.Client
	baseUrl   string
	appKey    string
	appSecret string
}

func NewSdk(baseUrl, appKey, appSecret string) *Sdk {
	httpClient := http.Client{}

	client := Sdk{
		client:    &httpClient,
		baseUrl:   baseUrl,
		appKey:    appKey,
		appSecret: appSecret,
	}

	return &client
}

// GetDevices 获取被试品信息
func (c *Sdk) GetDevices(testStationId int64, qrcode string) error {
	url := fmt.Sprintf("%s/devices?test_station_id=%d&qrcode=%s", c.baseUrl, testStationId, qrcode)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	err = c.signRequest(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[Err] status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(string(content))

	return nil
}

// SaveTestData 保存试验数据
func (c *Sdk) SaveTestData(qrcode string, testDataList []map[string]interface{}) error {
	content, err := json.Marshal(&testDataList)
	if err != nil {
		return err
	}

	body := bytes.NewReader(content)

	url := fmt.Sprintf("%s/devices/%s/test_data", c.baseUrl, qrcode)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	err = c.signRequest(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[Err] status code: %d", resp.StatusCode)
	}

	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(string(content))

	return nil
}

// UploadImage 上传试验图片
func (c *Sdk) UploadImage(imagePath string) error {
	image, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer image.Close()

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, image)
	if err != nil {
		return err
	}

	writer.Close()

	url := fmt.Sprintf("%s/uploads", c.baseUrl)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	err = c.signRequest(req)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[Err] status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(string(content))

	return nil
}

// signRequest 签名
func (c *Sdk) signRequest(req *http.Request) error {
	var query string
	queryVal := req.URL.Query()
	if len(queryVal) > 0 {
		var keys []string
		for key := range queryVal {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		var queryArr []string
		for _, key := range keys {
			queryArr = append(queryArr, fmt.Sprintf("%s=%s", key, queryVal.Get(key)))
		}

		query = strings.Join(queryArr, "&")
	}

	var contentMD5 string
	switch req.Method {
	case http.MethodGet, http.MethodDelete:
	case http.MethodPost, http.MethodPut:
		if req.Body != nil {
			content, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}

			contentMD5, err = Md5(content)
			if err != nil {
				return err
			}

			// add body back
			req.Body = ioutil.NopCloser(bytes.NewBuffer(content))
		}
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	stringToSign := req.Method + "\n" +
		req.URL.Path + "\n" +
		query + "\n" +
		req.Header.Get("Content-Type") + "\n" +
		contentMD5 + "\n" +
		timestamp + "\n" +
		c.appSecret

	signature, err := HmacSha1([]byte(c.appKey), []byte(stringToSign))
	if err != nil {
		return err
	}

	req.Header.Add("X-AppKey", c.appKey)
	req.Header.Add("X-Timestamp", timestamp)
	req.Header.Add("X-Signature", signature)

	return nil
}

func Md5(data []byte) (string, error) {
	h := md5.New()
	h.Write(data)
	bs := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(bs), nil
}

func HmacSha1(key, data []byte) (string, error) {
	hm := hmac.New(sha1.New, key)
	hm.Write(data)
	bs := hm.Sum(nil)

	return base64.StdEncoding.EncodeToString(bs), nil
}
