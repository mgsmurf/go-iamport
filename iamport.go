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

// Payment 결제 정보
type Payment struct {
	ImpUID        string `json:"imp_uid"`
	MerchantUID   string `json:"merchant_uid"`
	PayMethod     string `json:"pay_method"`
	PGProvider    string `json:"pg_provider"`
	PGTID         string `json:"pg_tid"`
	ApplyNum      string `json:"apply_num"`
	CardName      string `json:"card_name"`
	CardQuota     int    `json:"card_quota"`
	VBankName     string `json:"vbank_name"`
	VBankNum      string `json:"vbank_num"`
	VBankHolder   string `json:"vbank_holder"`
	Name          string `json:"name"`
	Amount        int64  `json:"amount"`
	CancelAmount  string `json:"cancel_amount"`
	BuyerName     string `json:"buyer_name"`
	BuyerEmail    string `json:"buyer_email"`
	BuyerTel      string `json:"buyer_tel"`
	BuyerAddr     string `json:"buyer_addr"`
	BuyerPostCode string `json:"buyer_postcode"`
	CustomData    string `json:"custom_data"`
	UserAgent     string `json:"user_agent"`
	Status        string `json:"status"`
	PaidAt        int64  `json:"paid_at"`
	FailedAt      int64  `json:"failed_at"`
	CanceledAt    int64  `json:"canceled_at"`
	FailReason    string `json:"fail_reason"`
	CancelReason  string `json:"cancel_reason"`
	ReceiptURL    string `json:"receipt_url"`
}

func (cli *Client) authorization() (string, error) {
	now := time.Now()

	switch {
	case cli.AccessToken.Token == "",
		cli.AccessToken.Expired.IsZero(),
		!cli.AccessToken.Expired.Before(now):

		err := cli.GetToken()
		if err != nil {
			return "", err
		}
	}

	return cli.AccessToken.Token, nil
}

// GetPaymentImpUID imp_uid로 결제 정보 가져오기
//
// GET /payments/{imp_uid}
func (cli *Client) GetPaymentImpUID(iuid string) (Payment, error) {
	data := struct {
		Code     int     `json:"code"`
		Message  string  `json:"string"`
		Response Payment `json:"response"`
	}{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.iamport.kr/payments/%s", iuid), nil)
	if err != nil {
		return data.Response, err
	}

	auth, err := cli.authorization()
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Authorization", auth)

	res, err := cli.HTTP.Do(req)
	if err != nil {
		return data.Response, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return data.Response, errors.New("iamport: unauthorized")
	}

	if res.StatusCode == http.StatusNotFound {
		return data.Response, errors.New("iamport: invalid imp_uid")
	}

	if res.StatusCode != http.StatusOK {
		return data.Response, errors.New("iamport: unknown error")
	}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data.Response, err
	}

	if data.Code != 0 {
		return data.Response, fmt.Errorf("iamport: %s", data.Message)
	}

	return data.Response, nil
}

// GetPaymentMerchantUID merchant_uid로 결제 정보 가져오기
//
// GET /payments/find/{merchant_uid}
func (cli *Client) GetPaymentMerchantUID(muid string) (Payment, error) {
	data := struct {
		Code     int     `json:"code"`
		Message  string  `json:"string"`
		Response Payment `json:"response"`
	}{}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.iamport.kr/payments/find/%s", muid), nil)
	if err != nil {
		return data.Response, err
	}

	auth, err := cli.authorization()
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Authorization", auth)

	res, err := cli.HTTP.Do(req)
	if err != nil {
		return data.Response, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return data.Response, errors.New("iamport: unauthorized")
	}

	if res.StatusCode == http.StatusNotFound {
		return data.Response, errors.New("iamport: invalid merchant_uid")
	}

	if res.StatusCode != http.StatusOK {
		return data.Response, errors.New("iamport: unknown error")
	}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data.Response, err
	}

	if data.Code != 0 {
		return data.Response, fmt.Errorf("iamport: %s", data.Message)
	}

	return data.Response, nil
}

// Status 결제 상태
type Status string

const (
	// StatusAll 전체
	StatusAll = "all"
	// StatusReady 미결제
	StatusReady = "ready"
	// StatusPaid 결제 완료
	StatusPaid = "paid"
	// StatusCanceled 결제 취소
	StatusCanceled = "canceled"
	// StatusFailed 결제 실패
	StatusFailed = "failed"
)

// PagedPayments 다수의 결제 정보
type PagedPayments struct {
	Total    int       `json:"total"`
	Previous int       `json:"previous"`
	Next     int       `json:"next"`
	Payments []Payment `json:"list"`
}

// GetPaymentsStatus 결제 상태에 따른 결제 정보들 가져오기
//
// GET /payments/status/{payment_status}
func (cli *Client) GetPaymentsStatus(status Status, page int) (PagedPayments, error) {
	data := struct {
		Code     int           `json:"code"`
		Message  string        `json:"string"`
		Response PagedPayments `json:"response"`
	}{}

	url := fmt.Sprintf("https://api.iamport.kr/payments/status/%s", status)
	if page > 0 {
		url += fmt.Sprintf("?page=%d", page)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {

		return data.Response, err
	}

	auth, err := cli.authorization()
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Authorization", auth)

	res, err := cli.HTTP.Do(req)
	if err != nil {
		return data.Response, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return data.Response, errors.New("iamport: unauthorized")
	}

	if res.StatusCode == http.StatusNotFound {
		return data.Response, errors.New("iamport: invalid status or page")
	}

	if res.StatusCode != http.StatusOK {
		return data.Response, errors.New("iamport: unknown error")
	}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return data.Response, err
	}

	if data.Code != 0 {
		return data.Response, fmt.Errorf("iamport: %s", data.Message)
	}

	return data.Response, nil
}
