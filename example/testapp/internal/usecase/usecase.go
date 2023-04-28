package usecase

type myService interface {
	// TODO: define interface to inject a service
}

type UseCase struct {
	myService myService
}

func NewUseCase(myService myService) *UseCase {
	return &UseCase{
		myService: myService,
	}
}
