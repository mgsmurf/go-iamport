package iamport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client 아임포트 서버와 통신하는 클라이언트
type Client struct {
	APIKey      string
	APISecret   string
	AccessToken accessToken
	HTTP        *http.Client
}

type accessToken struct {
	Token   string
	Expired time.Time
}

// NewClient 아임포트로 부터 부여받은 API Key와 Api Secret을 사용하여 클라이언트를 새로 만든다.
func NewClient(apiKey string, apiSecret string, cli *http.Client) *Client {
	if cli == nil {
		cli = &http.Client{}
	}

	return &Client{
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTP:      cli,
	}
}

// GetToken APIKey와 APISecret을 사용하여 AccessToken을 받아 온다.
func (cli *Client) GetToken() error {
	if cli.APIKey == "" {
		return errors.New("iamport: APIKey is missing")
	}

	if cli.APISecret == "" {
		return errors.New("iamport: APISecret is missing")
	}

	form := url.Values{}
	form.Set("imp_key", cli.APIKey)
	form.Set("imp_secret", cli.APISecret)

	res, err := cli.HTTP.PostForm("https://api.iamport.kr/users/getToken", form)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return errors.New("iamport: unauthorized")
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("iamport: unknown error")
	}

	data := struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Response struct {
			AccessToken string `json:"access_token"`
			ExpiredAt   int64  `json:"expired_at"`
			Now         int64  `json:"now"`
		} `json:"response"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return err
	}

	if data.Code != 0 {
		return fmt.Errorf("iamport: %s", data.Message)
	}

	cli.AccessToken.Token = data.Response.AccessToken
	cli.AccessToken.Expired = time.Unix(data.Response.ExpiredAt, 0)

	return nil
}
