package usecase

type myService interface {
	// TODO: define interface to inject a service
}

type pinPongalka interface {
	// TODO: define interface to inject a service
}

type UseCase struct {
	myService   myService
	pinPongalka pinPongalka
}

func NewUseCase(myService myService, pinPongalka pinPongalka) *UseCase {
	return &UseCase{
		myService:   myService,
		pinPongalka: pinPongalka,
	}
}
