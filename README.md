# go-iamport

Go Language 아임포트 Rest API Client  
https://api.iamport.kr

## 설치

    $ go get github.com/mgsmurf/go-iamport

## 예제
    client := &http.Client{} // 상황에 맞는 클라이언트 사용
    iam := iamport.NewClient("<your_api_key>", "<your_api_secret>", client)
    pay, err := iam.GetPaymentImpUID("<some imp_uid>")
    if err != nil {
      fmt.Println(err)
      return
    }

    fmt.Println(pay.Amount)
    fmt.Println(pay.MerchantUID)

### App Engine
    client := urlfetch.Client(ctx)
    iam := iamport.NewClient("<your_api_key>", "<your_api_secret>", client)
    ...

## 구현되어있는 기능 - https://api.iamport.kr

- authenticate
  - POST /users/getToken
- payments  
  - GET /payments/{imp_uid}
  - GET /payments/find/{merchant_uid}
  - GET /payments/status/{payment_status}
  - POST /payments/cancel
- payments.validation
  - POST /payments/prepare
  - GET /payments/prepare/{merchant_uid}

### 미구현

- subscribe
  - POST /subscribe/payments/onetime
  - POST /subscribe/payments/again
  - POST /subscribe/payments/schedule
  - POST /subscribe/payments/unschedule
- subscribe.customer
  - DELETE /subscribe/customers/{customer_uid}
  - GET /subscribe/customers/{customer_uid}
  - POST /subscribe/customers/{customer_uid}
