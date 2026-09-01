package services

import (
	"context"
	"strings"
	"test/internal/domain"
	"test/internal/errs"
	"test/internal/store"
	"time"
)

type KitService struct {
	store     *store.Store
	kitClient KitClient
}

func NewKitService(store *store.Store, kitClient KitClient) *KitService {
	return &KitService{
		store:     store,
		kitClient: kitClient,
	}
}

type KitClient interface {
	GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (string, error)
}

func separateData(s string) []string {
	item := strings.Split(s, ";")

	var data []string
	for i := range len(item) - 1 {
		data = append(data, item[i])
	}

	return data
}

func parseTime(s string) (time.Time, time.Time, error) {
	sep := strings.Split(s, "--")
	const lay = "2006-01-02 15:04:05"
	loc := time.Local
	t1, err := time.ParseInLocation(lay, sep[0], loc)
	if err != nil {
		return time.Time{}, time.Time{}, errs.Wrap(err)
	}
	t2, err := time.ParseInLocation(lay, sep[1], loc)
	if err != nil {
		return time.Time{}, time.Time{}, errs.Wrap(err)
	}
	return t1, t2, nil
}

func (a *KitService) GetCustomerByID(ctx context.Context, id int64) (domain.Customer, error) {
	var c domain.Customer
	c, err := a.store.GetCustomerDataByID(ctx, id)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (a *KitService) GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (string, error) {
	return a.kitClient.GetDataKitVending(ctx, companyId, userLogin, password, date)
}

func (a *KitService) CheckNewSale(ctx context.Context, c domain.Customer, drinks, date string) ([]string, error) {
	newDrinks := separateData(drinks)
	tStart, tEnd, err := parseTime(date)
	if err != nil {
		return nil, err
	}

	exists, err := a.store.ExistsStartDate(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	if !exists {
		return newDrinks, nil
	}

	oldDrinks, err := a.store.GetDrinks(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	if len(newDrinks) <= len(oldDrinks) {
		return nil, nil
	}

	end, err := a.store.GetEndDate(ctx, c.INN, tStart)
	if err != nil {
		return nil, err
	}

	date = end.Format("2006-01-02 15:04:05") + "--" + tEnd.Format("2006-01-02 15:04:05")
	newData, err := a.kitClient.GetDataKitVending(ctx, c.CompanyID, c.UserLogin, c.PasswordKit, date)
	if err != nil {
		return nil, err
	}
	return separateData(newData), nil
}

func (a *KitService) AddDrinks(ctx context.Context, c domain.Customer, drinks []string, date string) error {
	tStart, tEnd, err := parseTime(date)
	if err != nil {
		return err
	}

	var titleDrinks []string

	for i := range drinks {
		sep := strings.Split(drinks[i], ":")
		titleDrinks = append(titleDrinks, sep[0])
	}
	if err := a.store.InsertDrinks(ctx, c.INN, titleDrinks, tStart, tEnd); err != nil {
		return err
	}
	return nil
}
