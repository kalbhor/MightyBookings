package main

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	min = 100000000000
	max = 999999999999

	timeoutDuration = "300"
	currency        = "INR"
	description     = "Stuff"
	city            = "Manipal"
	country         = "India"
	zip             = "576104"
	gateway         = "https://biz.traknpay.in/v2/paymentrequest"
	returnURL       = "http://54.163.75.233:8000/paymentresponse" // callback url for traknpay
	failureURL      = "http://54.163.75.233:8000/failure"
)

var (
	paymentKey  = os.Getenv("PAYMENT_KEY")
	paymentSalt = os.Getenv("PAYMENT_SALT")
)

func CallPayment(m map[string]string) (*http.Response, error) {

	rand.Seed(time.Now().Unix())

	params := make(map[string]string)

	params["name"] = m["name"]
	params["amount"] = m["price"]
	params["email"] = m["email"]
	params["phone"] = m["phone"]

	params["quantity"] = m["quantity"]
	params["showid"] = m["showid"]
	params["category"] = m["category"]
	params["date"] = m["date"]
	params["time"] = m["time"]
	params["description"] = m["hall"]

	/*	params["hall"] = m["hall"]   // Redundant
		params["movie"] = m["movie"] // Redundant
	*/

	params["return_url_failure"] = failureURL
	params["return_url_cancel"] = failureURL
	params["timeout_duration"] = timeoutDuration
	params["currency"] = currency
	params["country"] = country
	params["city"] = city
	params["zip_code"] = zip
	params["return_url"] = returnURL
	params["order_id"] = m["order_id"] //
	params["api_key"] = paymentKey
	params["mode"] = "TEST" //

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hashData := paymentSalt
	for _, k := range keys {
		if len(params[k]) != 0 {
			hashData = hashData + "|" + params[k]
		}
	}

	sha_512 := sha512.New()
	sha_512.Write([]byte(hashData))

	hash := sha_512.Sum(nil)
	hashString := fmt.Sprintf("%x", hash)
	hashString = strings.ToUpper(hashString)

	params["hash"] = hashString

	v := url.Values{}

	for key := range params {
		v.Set(key, params[key])
	}
	resp, err := http.PostForm(gateway, v)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
