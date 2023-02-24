package config

type UseCase struct {
	Methods []string `yaml:"methods"`
	Deps    []string `yaml:"deps"`
}

func (u UseCase) GetConstructorName() string {
	return "NewUseCase"
}
