package myservice

import (
	"testapp/internal/service"
)

type myRepo interface {
	service.Transactor
	// TODO: define interface to inject a service or an adapter
}

type Service struct {
	myRepo myRepo
}

func NewService(myRepo myRepo) *Service {
	return &Service{

		myRepo: myRepo,
	}
}
