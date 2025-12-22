package store

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"test/internal/domain"
	"time"
)

/*type Customer struct {
	IdCustomer int64 `json:"IdCustomer"`
	CompanyId    string     `json:"CompanyId"`
	UserLogin    string     `json:"UserLogin"`
	PasswordKit  string     `json:"PasswordKit"`
	INN          string     `json:"INN"`
	PasswordFns  string     `json:"PasswordFns"`
	Date         string     `json:"Date"`
	Device       DeviceInfo `json:"Device"`
	RefreshToken string     `json:"RefreshToken"`
	Token        string     `json:"TokenType"`
}*/

type CustomerData struct {
	IDCustomer  int64
	CompanyId   string
	UserLogin   string
	PasswordKit string
	INN         string
	PasswordFns string
	AutoMode    bool
	IsPaid      bool
}

type CustomerDevice struct {
	JWT      string
	DeviceID string
}

func (s *Store) GetCustomerDataByInn(ctx context.Context, inn, companyId, userLogin string) (CustomerData, error) {
	var cd CustomerData
	err := s.Pool.QueryRow(ctx, `
		SELECT id, inn, company_id, user_login, password_kit, password_fns
		FROM customers_data
		WHERE inn = @inn AND company_id = @company_id AND user_login = @user_login
`,
		pgx.NamedArgs{
			"inn":        inn,
			"company_id": companyId,
			"user_login": userLogin,
		}).Scan(&cd.IDCustomer, &cd.INN, &cd.CompanyId, &cd.UserLogin, &cd.PasswordKit, &cd.PasswordFns)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CustomerData{}, nil
		}
		return CustomerData{}, err
	}
	return cd, nil
}

func (s *Store) GetCustomerDeviceId(ctx context.Context, idCustomer int64) (CustomerDevice, error) {
	var d CustomerDevice
	err := s.Pool.QueryRow(ctx, `
		SELECT jwt, device_id
		FROM customers_device
		WHERE customers_data_id = @customers_data_id
`,
		pgx.NamedArgs{
			"customers_data_id": idCustomer,
		}).Scan(&d.JWT, &d.DeviceID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CustomerDevice{}, nil
		}
		return CustomerDevice{}, err
	}
	return d, nil
}

func (s *Store) InsertCustomerData(ctx context.Context, customer CustomerData, q Querier) (int64, error) {
	var id int64
	err := q.QueryRow(ctx, `
		INSERT INTO customers_data (inn, company_id, user_login, password_kit, password_fns, auto_mode, is_paid)
		VALUES (@inn, @company_id, @user_login, @password_kit, @password_fns, @auto_mode, @is_paid)
		RETURNING id
`,
		pgx.NamedArgs{
			"inn":          customer.INN,
			"company_id":   customer.CompanyId,
			"user_login":   customer.UserLogin,
			"password_kit": customer.PasswordKit,
			"password_fns": customer.PasswordFns,
			"auto_mode":    customer.AutoMode,
			"is_paid":      customer.IsPaid,
		}).Scan(&id)
	return id, err
}

func (s *Store) InsertCustomerDevice(ctx context.Context, d CustomerDevice, refreshToken, token string, q Querier, id int64) error {
	_, err := q.Exec(ctx, `
			INSERT INTO customers_device(jwt, customers_data_id, device_id, refresh_token, token, last_token)
			VALUES (@jwt, @customers_data_id, @device_id, @refresh_token, @token, @last_token)
`,
		pgx.NamedArgs{
			"jwt":               d.JWT,
			"customers_data_id": id,
			"device_id":         d.DeviceID,
			"refresh_token":     refreshToken,
			"token":             token,
			"last_token":        time.Now(),
		})
	return err
}

// альтернатива pgx.CollectOneRow маппит данные из бд сразу в структуру
func (s *Store) GetCustomerDataByID(ctx context.Context, id int64) (domain.Customer, error) {
	c := domain.Customer{
		DeviceInfo: &domain.DeviceInfo{},
	}
	err := s.Pool.QueryRow(ctx, `
		SELECT inn, company_id, user_login, password_kit, password_fns, device_id, refresh_token, token
		FROM customers_data 
		INNER JOIN customers_device
		ON customers_data.id = customers_device.customers_data_id
		WHERE customers_data.id = @id
`,
		pgx.NamedArgs{
			"id": id,
		}).Scan(&c.INN, &c.CompanyID, &c.UserLogin, &c.PasswordKit, &c.PasswordFns, &c.DeviceInfo.SourceDeviceID, &c.RefreshToken, &c.Token)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Customer{}, nil
		}
		return domain.Customer{}, err
	}
	return c, nil
}

func (s *Store) ExistsStartDate(ctx context.Context, inn string, timeStart time.Time) (bool, error) {
	var existsStartDate bool
	err := s.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM customer_request_period
			WHERE inn = @inn
			AND started_at = @time_start
)
`,
		pgx.NamedArgs{
			"inn":        inn,
			"time_start": timeStart,
		}).Scan(&existsStartDate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return existsStartDate, nil
}

func (s *Store) GetDrinks(ctx context.Context, inn string, timeStart time.Time) ([]string, error) {
	var drinks []string
	err := s.Pool.QueryRow(ctx, `
		SELECT drinks
		FROM customer_request_period
		WHERE inn = @inn
		AND started_at = @time_start
`,
		pgx.NamedArgs{
			"inn":        inn,
			"time_start": timeStart,
		}).Scan(&drinks)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return drinks, nil
}

func (s *Store) GetEndDate(ctx context.Context, inn string, timeStart time.Time) (time.Time, error) {
	var end time.Time
	err := s.Pool.QueryRow(ctx, `
		SELECT ended_at
		FROM customer_request_period
		WHERE inn = @inn
		AND started_at = @time_start
`,
		pgx.NamedArgs{
			"inn":        inn,
			"time_start": timeStart,
		}).Scan(&end)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return end, nil
}

func (s *Store) InsertDrinks(ctx context.Context, inn string, drinks []string, start, end time.Time) error {
	_, err := s.Pool.Exec(ctx, `
			INSERT INTO customer_request_period (inn, started_at, ended_at, drinks)
			VALUES (@inn, @started_at, @ended_at, @drinks)
			ON CONFLICT (inn, started_at) DO UPDATE
			SET ended_at = EXCLUDED.ended_at,
				drinks = COALESCE(@drinks, ARRAY[]::text[]) || EXCLUDED.drinks,                                 
`,
		pgx.NamedArgs{
			"inn":        inn,
			"started_at": start,
			"ended_at":   end,
			"drinks":     drinks,
		})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}
