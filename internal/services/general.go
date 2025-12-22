package services

import (
	"log/slog"
	"os"
	"path/filepath"
)

var (
	Jobs        = NewCancelMap()
	jwtSecret   = []byte("Don't Be a Menace to South Central While Drinking Your Juice in the Hood")
	WorkingAuto = make(map[string]int32)
)

/*func AddCustomer(r Request, refreshToken, token string) string {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var customersDataID int64
	err = conn.QueryRow(ctx, `
		INSERT INTO customers_data (inn, company_id, user_login, password_kit, password_fns, auto_mode, is_paid)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
`, r.Data.INN, r.Data.CompanyId, r.Data.UserLogin, r.Data.PasswordKit, r.Data.PasswordFns, false, false).Scan(&customersDataID)
	if err != nil {
		fmt.Println("customers_data: ", err)
	}

	jwtToken := newToken(customersDataID)

	_, err = conn.Exec(ctx, `
		INSERT INTO customers_device(
		          jwt,customers_data_id, device_id,refresh_token,token,last_token
		)
		VALUES ($1,$2,$3,$4,$5,$6)
`, jwtToken, customersDataID, r.Device.SourceDeviceID, refreshToken, token, time.Now())
	if err != nil {
		fmt.Println("customers_device: ", err)
	}
	return jwtToken
}*/

/*func InitData(ctx context.Context, r SingleRequest, id int64) Customer {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var c Customer
	err = conn.QueryRow(ctx, `
		SELECT
			inn, company_id, user_login, password_kit, password_fns
		FROM customers_data
		WHERE id = $1
`, id).Scan(&c.INN, &c.CompanyId, &c.UserLogin, &c.PasswordKit, &c.PasswordFns)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("customers_data: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}

	err = conn.QueryRow(ctx, `
		SELECT
			device_id, refresh_token, token
		FROM customers_device
		WHERE customers_data_id = $1
`, id).Scan(&c.Device.SourceDeviceID, &c.RefreshToken, &c.Token)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("customers_data: в БД нет такой комбинации")
		}
		fmt.Println(err)
	}
	c.Date = r.Date
	return c
}*/

/*func AddDrinksDb(ctx context.Context, c Customer, drinks []string) {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var d []string
	for i := range drinks {
		sep := strings.Split(drinks[i], ":")
		d = append(d, sep[0])
	}
	tStart, tEnd := parseTime(c.Date)

	tag, err := conn.Exec(ctx, `
		INSERT INTO customer_request_period (inn,started_at, ended_at, drinks)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (inn,started_at) DO NOTHING;
`, c.INN, tStart, tEnd, d)
	if err != nil {
		fmt.Println(err)
	}

	if tag.RowsAffected() == 0 {
		_, err = conn.Exec(ctx, `
					UPDATE customer_request_period
					SET
					    drinks = COALESCE(drinks, ARRAY[]::text[]) || $3,
						ended_at = $4
					WHERE inn = $2
					AND started_at = $1
	`, tStart, c.INN, d, tEnd)
		if err != nil {
			fmt.Println(err)
		}
	}
}*/

/*func FnsTokenUpdate(ctx context.Context, c Customer, id int64) {
	ctx, conn, err := newPGConn()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close(ctx)

	var lastToken time.Time
	err = conn.QueryRow(ctx, `
					SELECT last_token
					FROM customers_device
					WHERE customers_data_id = $1;
		`, id).Scan(&lastToken)
	if errors.Is(err, pgx.ErrNoRows) {
		fmt.Println("customer not exist")
	}
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()
	if lastToken.Before(now.Add(-1 * time.Hour)) {
		_, err = conn.Exec(ctx, `
						UPDATE customers_device
						SET
						    last_token = $3,
							token = $2
						WHERE customers_data_id = $1;
		`, id, c.Token, now)
	}
	if err != nil {
		fmt.Println(err)
	}
}*/

func SlogLogger() {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				src := a.Value.Any().(*slog.Source) // value is *Source when AddSource=true
				src.File = filepath.Base(src.File)  // например, "general.go" вместо полного пути
				return slog.Any(a.Key, src)
			}
			if a.Key == slog.TimeKey && len(groups) == 0 {
				t := a.Value.Time().Local() // локальная зона
				a.Value = slog.StringValue(t.Format("2006-01-02T15:04:05"))
			}
			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	slog.Info("server started")
}
