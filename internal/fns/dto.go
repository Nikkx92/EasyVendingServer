package fns

import (
	"test/internal/domain"
)

type refreshTokenRequest struct {
	DeviceInfo domain.DeviceInfo `json:"deviceInfo"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
}

type tokensResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	RefreshToken string `json:"refreshToken"`
	Token        string `json:"token"`
}

type tokenRequest struct {
	DeviceInfo   domain.DeviceInfo `json:"deviceInfo"`
	RefreshToken string            `json:"refreshToken"`
}

type service struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
}

type clientFns struct {
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
	Client                          clientFns `json:"client"`
	PaymentType                     string    `json:"paymentType"`
	IgnoreMaxTotalIncomeRestriction bool      `json:"ignoreMaxTotalIncomeRestriction"`
}
