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
