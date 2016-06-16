package iamport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	CardQuota     int64  `json:"card_quota"`
	VBankName     string `json:"vbank_name"`
	VBankNum      string `json:"vbank_num"`
	VBankHolder   string `json:"vbank_holder"`
	VBankDate     int64  `json:"vbank_date"`
	Name          string `json:"name"`
	Amount        int64  `json:"amount"`
	CancelAmount  int64  `json:"cancel_amount"`
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
		Message  string  `json:"message"`
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
		Message  string  `json:"message"`
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
		Message  string        `json:"message"`
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

// Bank 은행코드
type Bank string

const (
	// Bank기업은행 기업은행
	Bank기업은행 = Bank("03")
	// Bank국민은행 국민은행
	Bank국민은행 = Bank("04")
	// Bank외환은행 외환은행
	Bank외환은행 = Bank("05")
	// Bank수협중앙회 수협중앙회
	Bank수협중앙회 = Bank("07")
	// Bank농협중앙회 농협중앙회
	Bank농협중앙회 = Bank("11")
	// Bank우리은행 우리은행
	Bank우리은행 = Bank("20")
	// BankSC제일은행 SC제일은행
	BankSC제일은행 = Bank("23")
	// Bank대구은행 대구은행
	Bank대구은행 = Bank("31")
	// Bank부산은행 부산은행
	Bank부산은행 = Bank("32")
	// Bank광주은행 광주은행
	Bank광주은행 = Bank("34")
	// Bank전북은행 전북은행
	Bank전북은행 = Bank("37")
	// Bank경남은행 경남은행
	Bank경남은행 = Bank("39")
	// Bank한국씨티은행 한국씨티은행
	Bank한국씨티은행 = Bank("53")
	// Bank우체국 우체국
	Bank우체국 = Bank("71")
	// Bank하나은행 하나은행
	Bank하나은행 = Bank("81")
	// Bank통합신한은행 통합신한은행
	Bank통합신한은행 = Bank("88")
	// Bank동양종합금융증권 동양종합금융증권
	Bank동양종합금융증권 = Bank("D1")
	// Bank현대증권 현대증권
	Bank현대증권 = Bank("D2")
	// Bank미래에셋증권 미래에셋증권
	Bank미래에셋증권 = Bank("D3")
	// Bank한국투자증권 한국투자증권
	Bank한국투자증권 = Bank("D4")
	// Bank우리투자증권 우리투자증권
	Bank우리투자증권 = Bank("D5")
	// Bank하이투자증권 하이투자증권
	Bank하이투자증권 = Bank("D6")
	// BankHMC투자증권 HMC투자증권
	BankHMC투자증권 = Bank("D7")
	// BankSK증권 SK증권
	BankSK증권 = Bank("D8")
	// Bank대신증권 대신증권
	Bank대신증권 = Bank("D9")
	// Bank하나대투증권 하나대투증권
	Bank하나대투증권 = Bank("DA")
	// Bank굿모닝신한증권 굿모닝신한증권
	Bank굿모닝신한증권 = Bank("DB")
	// Bank동부증권 동부증권
	Bank동부증권 = Bank("DC")
	// Bank유진투자증권 유진투자증권
	Bank유진투자증권 = Bank("DE")
	// Bank신영증권 신영증권
	Bank신영증권 = Bank("DF")
)

// CancelOptions 결제 취소 옵션
type CancelOptions struct {
	Amount        string
	Reason        string
	RefundHolder  string
	RefundBank    Bank
	RefundAccount string
}

func (ops *CancelOptions) form() url.Values {
	vals := url.Values{}

	if ops.Amount != "" {
		vals.Set("amount", ops.Amount)
	}

	if ops.Reason != "" {
		vals.Set("reason", ops.Reason)
	}

	if ops.RefundHolder != "" {
		vals.Set("refund_holder", ops.RefundHolder)
	}

	if ops.RefundBank != "" {
		vals.Set("refund_bank", string(ops.RefundBank))
	}

	if ops.RefundAccount != "" {
		vals.Set("refund_account", ops.RefundAccount)
	}

	return vals
}

// CancelPaymentImpUID imp_uid로 결제 취소하기
//
// GET /payments/cancel
func (cli *Client) CancelPaymentImpUID(iuid string, options *CancelOptions) (Payment, error) {
	return cli.cancelPayment("imp_uid", iuid, options)
}

// CancelPaymentMerchantUID merchant_uid로 결제 취소하기
//
// GET /payments/cancel
func (cli *Client) CancelPaymentMerchantUID(muid string, options *CancelOptions) (Payment, error) {
	return cli.cancelPayment("merchant_uid", muid, options)
}

func (cli *Client) cancelPayment(key string, uid string, options *CancelOptions) (Payment, error) {
	data := struct {
		Code     int     `json:"code"`
		Message  string  `json:"message"`
		Response Payment `json:"response"`
	}{}

	var form url.Values
	if options != nil {
		form = options.form()
	} else {
		form = url.Values{}
	}

	form.Set(key, uid)

	req, err := http.NewRequest("POST",
		"https://api.iamport.kr/payments/cancel",
		bytes.NewBufferString(form.Encode()))
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

// Prepared 사전 등록된 결제 정보
type Prepared struct {
	MerchantUID string `json:"merchant_uid"`
	Amount      int64  `json:"amount"`
}

// PreparePayment 결제 정보 사전 등록하기
//
// POST /payments/prepare
func (cli *Client) PreparePayment(muid string, amount int64) (Prepared, error) {
	data := struct {
		Code     int      `json:"code"`
		Message  string   `json:"message"`
		Response Prepared `json:"response"`
	}{}

	form := url.Values{}
	form.Set("merchant_uid", muid)
	form.Set("amount", strconv.FormatInt(amount, 10))

	req, err := http.NewRequest("POST",
		"https://api.iamport.kr/payments/prepare",
		bytes.NewBufferString(form.Encode()))
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

// GetPreparedPayment 사전 등록된 결제 정보 보기
//
// GET /payments/prepare/{merchant_uid}
func (cli *Client) GetPreparedPayment(muid string) (Prepared, error) {
	data := struct {
		Code     int      `json:"code"`
		Message  string   `json:"message"`
		Response Prepared `json:"response"`
	}{}

	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.iamport.kr/payments/prepare/%s", muid), nil)
	if err != nil {
		return data.Response, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		fmt.Println(res.StatusCode)
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
