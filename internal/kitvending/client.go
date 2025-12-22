package kitvending

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	hc      *http.Client
	baseURL string
}

func NewClient(hc *http.Client, baseURL string) *Client {
	return &Client{hc: hc, baseURL: baseURL}
}

func hashing(c, p string) (h, u string) {
	uniqueNumb := strconv.FormatInt(time.Now().UnixNano(), 10)
	data := c + p + uniqueNumb
	hash := md5.Sum([]byte(data))
	sign := hex.EncodeToString(hash[:])
	return sign, uniqueNumb
}

func (c *Client) GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (int, string, error) {
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
	//url := "https://api2.kit-invest.ru/APIService.svc/GetSales"
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/APIService.svc/GetSales",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Println("timeout request in getDataKitVending")
		fmt.Println(err)
	}

	resp, err := c.hc.Do(req)
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

	return code, str, nil
}

func (c *Client) ValidLoginKit(ctx context.Context, companyId, userLogin, pass string) error {
	sign, uniqNum := hashing(companyId, pass)
	reque := request{
		Auth: auth{
			CompanyId: companyId,
			RequestId: uniqNum,
			UserLogin: userLogin,
			Sign:      sign,
		},
	}
	jsonData, err := json.Marshal(reque)
	if err != nil {
		slog.Error(err.Error(), "ИНН", userLogin)
		return err
	}
	//url := "https://api2.kit-invest.ru/APIService.svc/GetModems"
	//url := "http://192.168.1.88:8080/test"
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/APIService.svc/GetModems",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		slog.Error(err.Error(), "ИНН", userLogin)
		return err
	}

	//client := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.hc.Do(req)
	if err != nil {
		slog.Error(err.Error(), "ИНН", userLogin)
		return err
	}
	defer resp.Body.Close()

	var res response
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		slog.Error(err.Error(), "ИНН", userLogin)
		return err
	}
	if res.ErrorMessage != "" {
		return fmt.Errorf(res.ErrorMessage)
	}
	return nil
}
