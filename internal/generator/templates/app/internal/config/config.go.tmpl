package config

import (
	"github.com/spf13/viper"
	"time"
)

type AppConfig struct {
	Name    string
	Env     string
	Version string
}

type HTTPConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type ThirdPartyService struct {
	Url string
}

type SQLConfig struct {
	Host              string
	Port              string
	DBName            string
	User              string
	Password          string
	Timeout           time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	InterpolateParams bool
	Charset           string
	ParseTime         bool
	Timezone          string
	Collation         string
	ConnMaxLifetime   time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
}

type Config struct {
	App AppConfig

	// Server
	HTTP HTTPConfig

	MainDB SQLConfig

	Service1 ThirdPartyService
}

func NewConfig(filePath string) (Config, error) {
	viper.SetConfigFile(filePath)

	conf := Config{}
	if err := viper.ReadInConfig(); err != nil {
		return conf, err
	}

	viper.SetDefault("app.name", "testapp")
	viper.SetDefault("app.env", "dev")
	viper.SetDefault("app.version", "v1")
	viper.SetDefault("http.read_timeout", "1s")
	viper.SetDefault("http.write_timeout", "1s")
	conf = Config{
		App: AppConfig{
			Name:    viper.GetString("app.name"),
			Env:     viper.GetString("app.env"),
			Version: viper.GetString("app.version"),
		},

		HTTP: HTTPConfig{
			Host:         viper.GetString("http.host"),
			Port:         viper.GetString("http.port"),
			ReadTimeout:  viper.GetDuration("http.read_timeout"),
			WriteTimeout: viper.GetDuration("http.write_timeout"),
		},

		MainDB: SQLConfig{
			Host:              viper.GetString("main_db.host"),
			Port:              viper.GetString("main_db.port"),
			DBName:            viper.GetString("main_db.name"),
			User:              viper.GetString("main_db.user"),
			Password:          viper.GetString("main_db.password"),
			ReadTimeout:       viper.GetDuration("main_db.read_timeout"),
			WriteTimeout:      viper.GetDuration("main_db.write_timeout"),
			Timeout:           viper.GetDuration("main_db.timeout"),
			InterpolateParams: false,
			Charset:           "UTF-8",
			ParseTime:         true,
			Timezone:          "Europe/Moscow",
			Collation:         "",
		},
	}

	return conf, nil
}
