package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"strings"
	kitpb "test/api"
	"test/internal/domain"
	"test/internal/services"
)

type Server struct {
	kitpb.UnimplementedKitNalogServiceServer
	auth *services.AuthService
	kit  *services.KitService
	fns  *services.FnsService
}

func NewGRPCHandler(auth *services.AuthService, kit *services.KitService) *Server {
	return &Server{
		auth: auth,
		kit:  kit,
	}
}

func getId(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Info("metadata is not provided")
		//fmt.Println("metadata is not provided")
		//return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		slog.Info("authorization is not provided")
		//fmt.Println("authorization token is not provided")
		//return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	jwt := strings.TrimPrefix(vals[0], "Bearer ")
	userId := services.GetIdFromJwt(jwt)
	return userId
}

func (s *Server) Auth(ctx context.Context, req *kitpb.Request) (*kitpb.AuthResponse, error) {
	r := services.Request{
		Data: &domain.LoginData{
			CompanyId:   req.Data.CompanyId,
			UserLogin:   req.Data.UserLogin,
			PasswordKit: req.Data.PasswordKit,
			INN:         req.Data.Inn,
			PasswordFns: req.Data.PasswordFns,
		},
		Device: &domain.DeviceInfo{
			SourceDeviceID: req.Device.SourceDeviceId,
			SourceType:     req.Device.SourceType,
			AppVersion:     req.Device.AppVersion,
			MetaDetails: &domain.MetaDetails{
				UserAgent: req.Device.MetaDetails.UserAgent,
			},
		},
		DataKit: &req.DataKit,
	}

	jwtToken, err := s.auth.Authorization(ctx, r)
	if err != nil {
		return nil, err
	}

	if jwtToken != "" {
		return &kitpb.AuthResponse{IsValid: true, Message: jwtToken}, nil
	} else {
		//err := kitvending.ValidLoginKit(req.Data.CompanyId, req.Data.UserLogin, req.Data.PasswordKit)
		err := s.auth.ValidLoginKit(ctx, req.Data.CompanyId, req.Data.UserLogin, req.Data.PasswordKit)
		if err != nil {
			return nil, err
		}

		//tokens, err := fns.GetRefreshToken(r.Device, req.Data.Inn, req.Data.PasswordFns)
		rt, t, err := s.auth.GetRefreshToken(ctx, r.Device, req.Data.Inn, req.Data.PasswordFns)
		if err != nil {
			return nil, err
		}

		jwtToken, err = s.auth.AddCustomer(ctx, r, rt, t)
		if err != nil {
			return nil, err
		}
		return &kitpb.AuthResponse{IsValid: true, Message: jwtToken}, nil
	}
}

func (s *Server) SendToKitVend(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.ResponseKit, error) {

	userId := getId(ctx)
	/*r := services.SingleRequest{
		Date: req.Date,
	}*/
	//data := services.InitData(ctx, r, userId)
	c, err := s.kit.GetCustomerByID(ctx, userId)
	if err != nil {
		return nil, err
	}

	//dataVending, message := services.ToKitVending(ctx, customerData, req.Date)
	//code, data, err := kitvending.GetDataKitVending(ctx, c.CompanyID, c.UserLogin, c.PasswordKit, req.Date)
	code, data, err := s.kit.GetDataKitVending(ctx, c.CompanyID, c.UserLogin, c.PasswordKit, req.Date)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return &kitpb.ResponseKit{Data: nil, Message: data}, nil
	}

	newData, err := s.kit.CheckNewSale(ctx, c, data, req.Date)
	if err != nil {
		return nil, err
	}
	return &kitpb.ResponseKit{Data: newData}, nil
	/*if message != "" {
		return &kitpb.ResponseKit{Data: nil, Message: message}, nil
	} else {
		return &kitpb.ResponseKit{Data: d, Message: message}, nil
	}*/
}

func (s *Server) SendToFns(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.ResponseSingleString, error) {
	userId := getId(ctx)
	/*r := services.SingleRequest{
		Date:    req.Date,
		DataKit: req.DataKit,
	}*/
	//data := services.InitData(ctx, r, userId)
	c, err := s.auth.GetCustomerByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	//services.AddDrinksDb(ctx, data, req.DataKit)
	/*if err := s.auth.AddDrinks(ctx, c, req.DataKit, req.Date); err != nil {
		return nil, err
	}*/
	//message, modifiedCustomer := services.Check(ctx, data, req.DataKit)
	//refreshedCustomer, err := fns.SendSaleToFns(ctx, c, req.DataKit)
	refreshedCustomer, err := s.kit.SendSaleToFns(ctx, c, req.DataKit)
	fmt.Println(refreshedCustomer)
	if err != nil {
		return nil, err
	}
	//services.FnsTokenUpdate(ctx, modifiedCustomer, userId)
	return &kitpb.ResponseSingleString{Message: "Отправлено успешно"}, nil
}

/*func (s *Server) Start(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.ResponseSingleString, error) {
	userId := getId(ctx)
	r := services.SingleRequest{}
	data := services.InitData(ctx, r, userId)

	if _, exists := services.Jobs.Get(data); exists {
		return &kitpb.ResponseSingleString{Message: fmt.Sprintf("Автоматическая отправка для %s уже включена", data.INN)}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	services.Jobs.Set(data.INN, cancel)

	go services.BackgroundWork(ctx, data, userId)
	services.WorkingAuto[data.INN]++
	return &kitpb.ResponseSingleString{Message: fmt.Sprintf("Автоматическая отправка для %s включена успешно", data.INN)}, nil
}

func (s *Server) Stop(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.ResponseSingleString, error) {
	userId := getId(ctx)
	r := services.SingleRequest{}
	data := services.InitData(ctx, r, userId)

	cancel, exists := services.Jobs.Get(data.INN)
	if !exists {
		return &kitpb.ResponseSingleString{Message: fmt.Sprintf("ИНН %s не найден", data.INN)}, nil
	}
	cancel()
	services.Jobs.Delete(data.INN)
	delete(services.WorkingAuto, data.INN)
	return &kitpb.ResponseSingleString{Message: fmt.Sprintf("Автоматическая отправка для %s остановлена", data.INN)}, nil
}

func (s *Server) Sales(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.RespOuterMapSales, error) {
	userId := getId(ctx)
	r := services.SingleRequest{}
	data := services.InitData(ctx, r, userId)

	sales := services.GetSales(ctx, data)

	resp := &kitpb.RespOuterMapSales{
		Data: make(map[string]*kitpb.RespInnerMapSales),
	}
	for k, v := range sales {
		sep := strings.Split(k, " ")
		resp.Data[sep[0]] = &kitpb.RespInnerMapSales{
			Values: v,
		}
	}

	return resp, nil
}*/

func (s *Server) WorkingAutoMode(ctx context.Context, req *kitpb.SingleRequest) (*kitpb.RespInnerMapSales, error) {
	userId := getId(ctx)
	if userId != 35 {
		return nil, fmt.Errorf("wrong id")
	}
	return &kitpb.RespInnerMapSales{Values: services.WorkingAuto}, nil
}
