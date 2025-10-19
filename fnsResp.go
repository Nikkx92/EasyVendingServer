package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func check() string {

	payload := CheckRequest{
		OperationTime: time.Now().Format(time.RFC3339),
		RequestTime:   time.Now().Format(time.RFC3339),
		Services: []Service{
			{Name: "Латте 300мл",
				Amount:   110,
				Quantity: 1},
		},
		TotalAmount: "110",
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
		log.Fatalf("Ошибка при маршалинге JSON: %v", err)
	}

	url := "https://lknpd.nalog.ru/api/v1/income"
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	request.Header.Set("Authorization", "Bearer "+"eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJ7XCJhdXRoVHlwZVwiOlwiTEtGTFwiLFwiZmlkXCI6MTAwMTc3MDEyOTExLFwibG9naW5cIjpcIjEwMDE3NzAxMjkxMVwiLFwiaWRcIjoyNjU4NjIwMyxcImRldmljZUlkXCI6XCJGR2pkajczSmhkLWpqZERfZDg2MzJcIixcIm9wZXJhdG9ySWRcIjpudWxsLFwiY3N1ZFVzZXJuYW1lXCI6bnVsbCxcInNlZ21lbnRzXCI6W10sXCJ0b2tlbklkXCI6bnVsbCxcImZpcnN0U2VnbWVudFwiOm51bGx9IiwiZXhwIjoxNzU4NzkyODg0fQ.TBKUYSOJ5Yy0pjrNk8IctFz1QZuzv0C0FAnlpvHfLNvO34gtuJIZERhMmpBSvQbZaSwW4xCatYfQkGsuhM6qNQ")
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(request)
	defer resp.Body.Close()

	var checkResponse CheckResponse
	err = json.NewDecoder(resp.Body).Decode(&checkResponse)
	if err != nil {
		fmt.Println("Ошибка:", err)
	}
	return checkResponse.Code
}
