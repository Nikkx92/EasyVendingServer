package services

import (
	"context"
	"github.com/jackc/pgx/v5"
	"test/internal/domain"
	"test/internal/fns"
	"test/internal/jwtauth"
	"test/internal/kitvending"
	"test/internal/store"
)

type Request struct {
	Data    *domain.LoginData  `json:"Data"`
	Device  *domain.DeviceInfo `json:"Device"`
	DataKit *[]string          `json:"dataKit"`
}

type AuthResponse struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
}

type FNSAuth interface {
	GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error)
}

type KitAuth interface {
	ValidLoginKit(ctx context.Context, companyId, userLogin, pass string) error
}

type AuthService struct {
	store   *store.Store
	fnsAuth FNSAuth
	kitAuth KitAuth
}

func NewAuthService(st *store.Store, fnsC *fns.Client, kitC *kitvending.Client) *AuthService {
	return &AuthService{
		store:   st,
		fnsAuth: fnsC,
		kitAuth: kitC,
	}
}

/*
	func Authorization(ctx context.Context, r Request) (string, error) {
		conn, err := newPGConn(ctx)
		if err != nil {
			slog.Error(err.Error(), "ИНН", r.Data.INN)
			return "", fmt.Errorf("сервис бд не доступен, попробуйте позже")
		}
		defer conn.Close(ctx)

		var c Customer
		var idCustomer int64
		err = conn.QueryRow(ctx, `
			SELECT
				id, inn, company_id, user_login, password_kit, password_fns
			FROM customers_data
			WHERE inn = $1
				AND company_id = $2
				AND user_login = $3

`, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin).Scan(&idCustomer, &c.INN, &c.CompanyId, &c.UserLogin, &c.PasswordKit, &c.PasswordFns)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("в БД нет такой комбинации", "ИНН", r.Data.INN)
		}
		slog.Info("no rows in result set", "ИНН", r.Data.INN)
	}

	var token string
	err = conn.QueryRow(ctx, `
		SELECT
		    jwtauth, device_id
		FROM customers_device
		WHERE customers_data_id = $1

`, idCustomer).Scan(&token, &c.Device.SourceDeviceID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				slog.Info("в БД нет такой комбинации", "ИНН", r.Data.INN)
			}
			slog.Info("no rows in result set", "ИНН", r.Data.INN)
		}

		ok :=
			r.Data.INN == c.INN &&
				r.Data.CompanyId == c.CompanyId &&
				r.Data.UserLogin == c.UserLogin &&
				r.Data.PasswordKit == c.PasswordKit &&
				r.Data.PasswordFns == c.PasswordFns &&
				r.Device.SourceDeviceID == c.Device.SourceDeviceID

		if ok {
			return token, nil
		} else {
			return "", nil
		}
	}
*/

func (a *AuthService) Authorization(ctx context.Context, r Request) (string, error) {
	cd, err := a.store.GetCustomerDataByInn(ctx, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin)
	if err != nil {
		return "", err
	}

	dev, err := a.store.GetCustomerDeviceId(ctx, cd.IDCustomer)
	if err != nil {
		return "", err
	}

	ok :=
		r.Data.INN == cd.INN &&
			r.Data.CompanyId == cd.CompanyId &&
			r.Data.UserLogin == cd.UserLogin &&
			r.Data.PasswordKit == cd.PasswordKit &&
			r.Data.PasswordFns == cd.PasswordFns &&
			r.Device.SourceDeviceID == dev.DeviceID

	if !ok {
		return "", nil
	}
	return dev.JWT, nil
}

func (a *AuthService) AddCustomer(ctx context.Context, r Request, refreshToken, token string) (string, error) {
	var jwtToken string

	data := store.CustomerData{
		CompanyId:   r.Data.CompanyId,
		UserLogin:   r.Data.UserLogin,
		PasswordKit: r.Data.PasswordKit,
		INN:         r.Data.INN,
		PasswordFns: r.Data.PasswordFns,
		AutoMode:    false,
		IsPaid:      false,
	}

	err := pgx.BeginFunc(ctx, a.store.Pool, func(tx pgx.Tx) error {
		id, err := a.store.InsertCustomerData(ctx, data, tx)
		if err != nil {
			return err
		}

		jwtToken, err = jwtauth.NewJwt(id)
		if err != nil {
			return err
		}
		dev := store.CustomerDevice{
			JWT:      jwtToken,
			DeviceID: r.Device.SourceDeviceID,
		}

		if err := a.store.InsertCustomerDevice(ctx, dev, refreshToken, token, tx, id); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return jwtToken, nil
}

func (a *AuthService) ValidLoginKit(ctx context.Context, companyId, userLogin, pass string) error {
	return a.kitAuth.ValidLoginKit(ctx, companyId, userLogin, pass)
}

func (a *AuthService) GetRefreshToken(ctx context.Context, d *domain.DeviceInfo, inn, pass string) (string, string, error) {
	return a.fnsAuth.GetRefreshToken(ctx, d, inn, pass)
}
