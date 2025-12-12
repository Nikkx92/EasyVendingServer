package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type service struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
}

type client struct {
	ContactPhone *string `json:"contactPhone"`
	DisplayName  *string `json:"displayName"`
	Inn          *string `json:"inn"`
	IncomeType   string  `json:"incomeType"`
}

type checkRequest struct {
	OperationTime                   string    `json:"operationTime"`
	RequestTime                     string    `json:"requestTime"`
	Services                        []service `json:"services"`
	TotalAmount                     string    `json:"totalAmount"`
	Client                          client    `json:"client"`
	PaymentType                     string    `json:"paymentType"`
	IgnoreMaxTotalIncomeRestriction bool      `json:"ignoreMaxTotalIncomeRestriction"`
}

type checkResponse struct {
	Code string `json:"code"`
}

func Check(c Customer, drinks []string) (string, Customer) {
	duplicates := make(map[string]int)
	for _, s := range drinks {
		duplicates[s]++
	}

	const maxRetries = 1
	var st string
outer:
	for attempt := 0; attempt <= maxRetries; attempt++ {
		for i := range duplicates {
			sep := strings.Split(i, ":")
			n, _ := strconv.Atoi(sep[1])
			nf := float64(n)

			payload := checkRequest{
				OperationTime: time.Now().Format(time.RFC3339),
				RequestTime:   time.Now().Format(time.RFC3339),
				Services: []service{
					{Name: sep[0],
						Amount:   nf,
						Quantity: duplicates[i]},
				},
				TotalAmount: strconv.Itoa(n * duplicates[i]),
				Client: client{
					ContactPhone: nil,
					DisplayName:  nil,
					Inn:          nil,
					IncomeType:   "FROM_INDIVIDUAL",
				},
				PaymentType:                     "CASH",
				IgnoreMaxTotalIncomeRestriction: false,
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				//main.logger.Println(err, "check")
				log.Fatalf("Ошибка при маршалинге JSON: %v", err)
			}

			url := "https://lknpd.nalog.ru/api/v1/income"
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println(err)
				//main.logger.Println(err, "check")
			}

			req.Header.Set("Authorization", "Bearer "+c.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				//main.logger.Println(err, "check")
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				//main.logger.Println(err, "getRefreshToken")
			}

			var res checkResponse
			if err = json.Unmarshal(body, &res); err != nil {
				//main.logger.Println(err, "check")
			}

			st = res.Code
			if res.Code == "Not Authenticated" {
				refTokenOrMessage, tokenOrEmpty := getRefreshToken(c.Device, c.INN, c.PasswordFns)
				if tokenOrEmpty == "" {
					return refTokenOrMessage, c
				} else {
					c.RefreshToken = refTokenOrMessage
					c.Token = tokenOrEmpty
					continue outer
				}
			} else if res.Code == "authentication.failed.expired.token" {
				t := getToken(c.Device, c.RefreshToken)
				if t == "authentication.failed" {
					refTokenOrMessage, tokenOrEmpty := getRefreshToken(c.Device, c.INN, c.PasswordFns)
					if tokenOrEmpty == "" {
						return refTokenOrMessage, c
					} else {
						c.RefreshToken = refTokenOrMessage
						c.Token = tokenOrEmpty
						continue outer
					}
				} else {
					c.Token = t
					continue outer
				}

			} else {
				st = "Отправлено успешно"
			}
		}
		break
	}
	return st, c
	//return "Отправлено успешно", c
}
