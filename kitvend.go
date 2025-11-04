package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Auth struct {
	CompanyId string `json:"companyid"`
	RequestId string `json:"requestid"`
	UserLogin string `json:"userlogin"`
	Sign      string `json:"sign"`
}

type Filter struct {
	UpDate string `json:"update"`
	ToDate string `json:"todate"`
}

type RequestKit struct {
	Auth   Auth   `json:"Auth"`
	Filter Filter `json:"Filter"`
}

type ResponseKit struct {
	ErrorMessage string `json:"ErrorMessage"`
	ResultCode   int    `json:"ResultCode"`
	Sales        []Sale `json:"Sales"`
}

type Sale struct {
	GoodsName string  `json:"GoodsName"`
	Sum       float64 `json:"Sum"`
}

func hashing(c, p string) (h, u string) {
	uniqueNumb := strconv.FormatInt(time.Now().UnixNano(), 10)
	data := c + p + uniqueNumb
	hash := md5.Sum([]byte(data))
	sign := hex.EncodeToString(hash[:])
	return sign, uniqueNumb
}

func kitRequest(c, uN, uL, s, uD, tD string) (int, string) {
	requ := RequestKit{
		Auth: Auth{
			CompanyId: c,
			RequestId: uN,
			UserLogin: uL,
			Sign:      s,
		},
		Filter: Filter{
			UpDate: uD + " 00:00:00",
			ToDate: tD + " 23:59:00",
		},
	}

	jsonData, _ := json.Marshal(requ)
	url := "https://api2.kit-invest.ru/APIService.svc/GetSales"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Println(err, "kitRequest")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Println(err, "kitRequest")
	}
	defer resp.Body.Close()

	//body, _ := io.ReadAll(resp.Body)

	var response ResponseKit
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Println(err, "kitRequest")
	}

	if response.ResultCode == 0 {
		var str string
		for _, sale := range response.Sales {
			//fmt.Printf("Товар: %s, Цена: %.2f руб.\n",
			//sale.GoodsName, sale.Sum)
			str += sale.GoodsName + ":" + strconv.Itoa(int(sale.Sum)) + ";"
		}
		return response.ResultCode, str
	} else {
		return response.ResultCode, response.ErrorMessage
	}
}

func getDataKitVending(companyId, userLogin, password, date string) (int, string) {
	sign, uniqNum := hashing(companyId, password)
	matchInterval := regexp.MustCompile(`^\d\d.\d\d.\d\d-\d\d.\d\d.\d\d$`)
	var upDate string
	var toDate string
	if matchInterval.MatchString(date) {
		s := strings.Split(date, "-")
		upDate = s[0]
		toDate = s[1]
	} else {
		upDate = date
		toDate = date
	}

	codeResponse, textResponse := kitRequest(companyId, uniqNum, userLogin, sign, upDate, toDate)

	return codeResponse, textResponse
}
