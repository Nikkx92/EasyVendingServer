package services

import (
	"context"
	"strings"
	"test/internal/domain"
	"test/internal/store"
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
	GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (int, string, error)
}

func separateData(s string) []string {
	item := strings.Split(s, ";")
	/*var drink []string
	var price []string
	for i := 0; i < len(item)-1; i++ {
		d := strings.Split(item[i], ":")
		drink = append(drink, d[0])
		price = append(price, d[1])
	}*/
	var data []string
	for i := range len(item) - 1 {
		data = append(data, item[i])
	}

	return data
}

func (a *KitService) GetCustomerByID(ctx context.Context, id int64) (domain.Customer, error) {
	var c domain.Customer
	c, err := a.store.GetCustomerDataByID(ctx, id)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (a *KitService) GetDataKitVending(ctx context.Context, companyId, userLogin, password, date string) (int, string, error) {
	return a.kitClient.GetDataKitVending(ctx, companyId, userLogin, password, date)
}

func (a *KitService) CheckNewSale(ctx context.Context, c domain.Customer, drinks, date string) ([]string, error) {
	newDrinks := separateData(drinks)
	tStart, tEnd := parseTime(date)

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
	_, newData, err := a.kitClient.GetDataKitVending(ctx, c.CompanyID, c.UserLogin, c.PasswordKit, date)
	if err != nil {
		return nil, err
	}
	return separateData(newData), nil
}

func (a *KitService) AddDrinks(ctx context.Context, c domain.Customer, drinks []string, date string) error {
	tStart, tEnd := parseTime(date)
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
