// Code generated by goro; DO NOT EDIT.

package myservice

// This file was generated by the goro tool.
// Editing this file might prove futile when you re-run the goro commands

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
