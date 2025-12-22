package fns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"test/internal/domain"
	"time"
)

type Client struct {
	hc      *http.Client
	baseURL string
}

func NewClient(hc *http.Client, baseURL string) *Client {
	return &Client{hc: hc, baseURL: baseURL}
}

func (c *Client) GetToken(ctx context.Context, d *domain.DeviceInfo, rf string) (string, error) {
	payload := tokenRequest{
		DeviceInfo:   *d,
		RefreshToken: rf,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	//url := "https://lknpd.nalog.ru/api/v1/auth/token"

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/api/v1/auth/token",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	//client := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res tokensResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", &BusinessError{Message: res.Message}
	}

	return res.Token, nil

}

func (c *Client) GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error) {
	payload := refreshTokenRequest{
		DeviceInfo: *d,
		Username:   inn,
		Password:   pass,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		slog.Error(err.Error(), "ИНН", inn)
		return "", "", err
	}

	//url := "https://lknpd.nalog.ru/api/v1/auth/lkfl"

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.baseURL+"/api/v1/auth/lkfl",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		slog.Error(err.Error(), "ИНН", inn)
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/json")

	//client := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.hc.Do(req)
	if err != nil {
		slog.Error(err.Error(), "ИНН", inn)
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error(), "ИНН", inn)
		return "", "", err
	}

	var res tokensResponse
	if err = json.Unmarshal(body, &res); err != nil {
		slog.Error(err.Error(), "ИНН", inn)
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", &BusinessError{Message: res.Message}
	}

	return res.RefreshToken, res.Token, nil
}

func (c *Client) SendSaleToFns(ctx context.Context, cus domain.Customer, drinks []string) (domain.Customer, error) {
	duplicates := make(map[string]int)
	for _, s := range drinks {
		duplicates[s]++
	}

	const maxRetries = 1
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
				Client: clientFns{
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
				return domain.Customer{}, err
			}

			//url := "https://lknpd.nalog.ru/api/v1/income"
			req, err := http.NewRequestWithContext(
				ctx,
				"POST",
				c.baseURL+"/api/v1/income",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				return domain.Customer{}, err
			}

			req.Header.Set("Authorization", "Bearer "+cus.Token)
			req.Header.Set("Content-Type", "application/json")

			//client := &http.Client{}
			resp, err := c.hc.Do(req)
			if err != nil {
				return domain.Customer{}, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				token, err := c.GetToken(ctx, cus.DeviceInfo, cus.RefreshToken)
				if err != nil {
					var be *BusinessError
					if errors.As(err, &be) {
						refToken, tok, err := c.GetRefreshToken(ctx, cus.DeviceInfo, cus.INN, cus.PasswordFns)
						if err != nil {
							return domain.Customer{}, err
						}
						cus.RefreshToken = refToken
						cus.Token = tok
						continue outer
					}
					return domain.Customer{}, err
				}
				cus.Token = token
				continue outer
			}
		}
		break outer
	}
	return cus, nil
}
