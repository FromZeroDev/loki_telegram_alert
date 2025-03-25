package common

import "github.com/caarlos0/env/v10"

type Configuration struct {
	SendFronted bool `env:"SEND_FRONTEND" envDefault:"TRUE"`
}

var Config Configuration = Configuration{}

func Load() error {
	return env.Parse(&Config)
}
