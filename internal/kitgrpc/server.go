package kitgrpc

import (
	"context"
	kitpb "test/api"
	"test/internal/services"
)

type Server struct {
	kitpb.UnimplementedKitNalogServiceServer

	// сюда можно внедрить зависимости: БД, логгер и т.д.
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Auth(ctx context.Context, req *kitpb.Request) (*kitpb.AuthResponse, error) {
	// здесь можно переиспользовать твою authorization / validLoginKit / addCustomer
	r := services.Request{
		Data: &services.LoginData{
			CompanyId:   req.Data.CompanyId,
			UserLogin:   req.Data.UserLogin,
			PasswordKit: req.Data.PasswordKit,
			INN:         req.Data.Inn,
			PasswordFns: req.Data.PasswordFns,
		},
		Device: &services.DeviceInfo{
			SourceDeviceID: req.Device.SourceDeviceId,
			SourceType:     req.Device.SourceType,
			AppVersion:     req.Device.AppVersion,
			MetaDetails: &services.MetaDetails{
				UserAgent: req.Device.MetaDetails.UserAgent,
			},
		},
		DataKit: &req.DataKit,
	}
	exist, jwtToken := services.Authorization(r)
	if exist {
		return &kitpb.AuthResponse{IsValid: true, Message: jwtToken}, nil
	}
	// упрощённо:
	return &kitpb.AuthResponse{IsValid: false, Message: "invalid credentials"}, nil
}
