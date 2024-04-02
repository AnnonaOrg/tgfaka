package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopay/internal/exts/config"
	my_log "gopay/internal/exts/log"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func ClientProxy(client *http.Client, proxy config.Proxy) error {
	if config.GetSiteConfig().Proxy.EnableProxy == false {
		return nil
	}
	proxyString := fmt.Sprintf("%s://%s:%d", proxy.Protocol, proxy.Host, proxy.Port)
	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		return err
	}
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	return nil
}

func sendRequest(method string, url string, dataMap *map[string]interface{}, options ...interface{}) ([]byte, error) {
	var req *http.Request
	var err error
	var resp *http.Response
	var respBytes []byte

	//defer func() {
	//	if err != nil {
	//		var data string
	//		if respBytes != nil {
	//			data = string(respBytes)
	//		}
	//		msgText := fmt.Sprintf("url: %s, 数据: %v, 返回值: %s, 错误: %s", url, dataMap, data, err.Error())
	//		tg_bot.SendAdmin(msgText)
	//	}
	//}()

	if dataMap != nil {
		jsonData, err := json.Marshal(dataMap)
		if err != nil {
			return nil, err
		}
		data := bytes.NewBuffer(jsonData)
		req, err = http.NewRequest(method, url, data)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	// Set content type if body is not nil (applicable for POST requests)
	if dataMap != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}
	for _, value := range options {
		switch value := value.(type) {
		case config.Proxy:
			err = ClientProxy(&client, value)
			if err != nil {
				return nil, err
			}
		case config.Headers:
			for key, value := range value {
				req.Header.Set(key, value)
			}
		}
	}

	my_log.LogDebug(fmt.Sprintf("%s: %s ,Data: %d", method, url, dataMap))
	for i := 0; i < 1; i++ {
		resp, err = client.Do(req)
		if err != nil {
			if isTimeoutError(err) {
				time.Sleep(time.Second * 1)
				continue
			} else {
				return nil, err
			}
		}
		break
	}

	//resp, err = client.Do(req)
	//if err != nil {
	//	return nil, err
	//}

	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("状态码 %d", resp.StatusCode))
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		return nil, err
	}

	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return respBytes, err
	}
	return respBytes, nil
}

func Post(url string, dataMap map[string]interface{}, options ...interface{}) ([]byte, error) {
	proxy := config.GetSiteConfig().Proxy
	options = append(options, proxy)
	return sendRequest(MethodPost, url, &dataMap, options...)
}
func Get(url string, options ...interface{}) ([]byte, error) {
	proxy := config.GetSiteConfig().Proxy
	options = append(options, proxy)
	return sendRequest(MethodGet, url, nil, options...)
}

const (
	MethodPost = "POST"
	MethodGet  = "GET"
)

func isTimeoutError(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}
