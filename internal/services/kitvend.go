package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type auth struct {
	CompanyId string `json:"companyid"`
	RequestId string `json:"requestid"`
	UserLogin string `json:"userlogin"`
	Sign      string `json:"sign"`
}

type filter struct {
	UpDate string `json:"update"`
	ToDate string `json:"todate"`
}

type request struct {
	Auth   auth   `json:"Auth"`
	Filter filter `json:"Filter"`
}

type response struct {
	ErrorMessage string `json:"ErrorMessage"`
	ResultCode   int    `json:"ResultCode"`
	Sales        []sale `json:"Sales"`
}

type sale struct {
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

func getDataKitVending(companyId, userLogin, password string, date string) (int, string) {
	var str string
	var code int

	sign, uniqNum := hashing(companyId, password)
	var upDate string
	var toDate string
	sep := strings.Split(date, "--")
	upDate = sep[0]
	toDate = sep[1]

	requ := request{
		Auth: auth{
			CompanyId: companyId,
			RequestId: uniqNum,
			UserLogin: userLogin,
			Sign:      sign,
		},
		Filter: filter{
			UpDate: upDate,
			ToDate: toDate,
		},
	}

	jsonData, _ := json.Marshal(requ)
	url := "https://api2.kit-invest.ru/APIService.svc/GetSales"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		fmt.Println(err)
	}

	if r.ResultCode == 0 {
		for _, s := range r.Sales {
			//fmt.Printf("Товар: %s, Цена: %.2f руб.\n",
			//sale.GoodsName, sale.Sum)
			str += s.GoodsName + ":" + strconv.Itoa(int(s.Sum)) + ";"
		}
		code = r.ResultCode
	} else {
		str = r.ErrorMessage
		code = r.ResultCode
	}

	return code, str
}

func ValidLoginKit(r Request) string {
	sign, uniqNum := hashing(r.Data.CompanyId, r.Data.PasswordKit)
	reque := request{
		Auth: auth{
			CompanyId: r.Data.CompanyId,
			RequestId: uniqNum,
			UserLogin: r.Data.UserLogin,
			Sign:      sign,
		},
	}
	jsonData, _ := json.Marshal(reque)
	url := "https://api2.kit-invest.ru/APIService.svc/GetModems"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var res response
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		fmt.Println(err)
	}
	return res.ErrorMessage
}
