package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Service struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
}

type Client struct {
	ContactPhone *string `json:"contactPhone"`
	DisplayName  *string `json:"displayName"`
	Inn          *string `json:"inn"`
	IncomeType   string  `json:"incomeType"`
}

type CheckRequest struct { // Основная структура запроса
	OperationTime                   string    `json:"operationTime"`
	RequestTime                     string    `json:"requestTime"`
	Services                        []Service `json:"services"`
	TotalAmount                     string    `json:"totalAmount"`
	Client                          Client    `json:"client"`
	PaymentType                     string    `json:"paymentType"`
	IgnoreMaxTotalIncomeRestriction bool      `json:"ignoreMaxTotalIncomeRestriction"`
}

type CheckResponse struct {
	Code string `json:"code"`
}

func check(r Request, d, p []string) string {

	const maxRetries = 1
	var st string
	for attempt := 0; attempt <= maxRetries; attempt++ {
		for i := range d {
			n, _ := strconv.Atoi(p[i])
			nf := float64(n)
			payload := CheckRequest{
				OperationTime: time.Now().Format(time.RFC3339),
				RequestTime:   time.Now().Format(time.RFC3339),
				Services: []Service{
					{Name: d[i],
						Amount:   nf,
						Quantity: 1},
				},
				TotalAmount: p[i],
				Client: Client{
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
				logger.Println(err, "check")
				log.Fatalf("Ошибка при маршалинге JSON: %v", err)
			}

			url := "https://lknpd.nalog.ru/api/v1/income"
			request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				logger.Println(err, "check")
			}

			request.Header.Set("Authorization", "Bearer "+r.Data.Token)
			request.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(request)
			if err != nil {
				logger.Println(err, "check")
			}
			defer resp.Body.Close()

			/*body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
			*/
			var checkResponse CheckResponse
			err = json.NewDecoder(resp.Body).Decode(&checkResponse)
			if err != nil {
				logger.Println(err, "check")
			}

			switch checkResponse.Code {
			case "authentication.failed.expired.token":
				t := getToken(r.Device, r.Data.RefreshToken)
				if t != "authentication.failed" {
					r.Data.Token = t
					st = ""
					continue
				} else {
					st = t
				}
			}
		}
	}
	return st
}
