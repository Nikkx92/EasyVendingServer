package services

import (
	"context"
	"test/internal/domain"
	"test/internal/store"
)

type FnsClient interface {
	GetToken(ctx context.Context, d *domain.DeviceInfo, rf string) (string, error)
	SendSaleToFns(ctx context.Context, cus domain.Customer, drinks []string) (domain.Customer, error)
	GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error)
}

type FnsService struct {
	store     *store.Store
	fnsClient FnsClient
}

func NewFnsService(store *store.Store, fnsClient FnsClient) *FnsService {
	return &FnsService{
		store:     store,
		fnsClient: fnsClient,
	}
}

/*func SendSaleToFns(ctx context.Context, c domain.Customer, drinks []string) (domain.Customer, error) {
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
				return domain.Customer{}, errTechnic
			}

			url := "https://lknpd.nalog.ru/api/v1/income"
			req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				return domain.Customer{}, errTechnic
			}

			req.Header.Set("Authorization", "Bearer "+c.Token)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return domain.Customer{}, errTechnic
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				token, err := GetToken(c.DeviceInfo, c.RefreshToken)
				if err != nil {
					var be *BusinessError
					if errors.As(err, &be) {
						tokens, err := GetRefreshToken(c.DeviceInfo, c.INN, c.PasswordFns)
						if err != nil {
							return domain.Customer{}, err
						}
						c.RefreshToken = tokens.RefreshToken
						c.Token = tokens.Token
						continue outer
					}
					return domain.Customer{}, err
				}
				c.Token = token
				continue outer
			}
		}
		break outer
	}
	return c, nil
}*/
