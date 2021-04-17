package config

import "fmt"

type Env struct {
	RestEndpoint string `yaml:"RestEndpoint"`
	WsEndpoint   string `yaml:"WsEndpoint"`
	IsSimulation bool   `yaml:"IsSimulation"`
}

type ApiInfo struct {
	ApiKey     string `yaml:"ApiKey"`
	SecretKey  string `yaml:"SecretKey"`
	Passphrase string `yaml:"Passphrase"`
}

type MetaData struct {
	Description string `yaml:"Description"`
}

type Config struct {
	MetaData `yaml:"MetaData"`
	Env      `yaml:"Env"`
	ApiInfo  `yaml:"ApiInfo"`
}

func (s *ApiInfo) String() string {
	res := "ApiInfo{"
	res += fmt.Sprintf("ApiKey:%v,SecretKey:%v,Passphrase:%v", s.ApiKey, s.SecretKey, s.Passphrase)
	res += "}"
	return res
}
