package iamport

import (
	"net/http"
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
