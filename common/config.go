package common

type Configuration struct{}

var Config Configuration = Configuration{}

func Load() error {
	return nil
}
