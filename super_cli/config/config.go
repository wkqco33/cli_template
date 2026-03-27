package config

// Config 애플리케이션 설정 구조체
type Config struct {
	AppName string `wconf:"NAME" default:"super_cli"`
	Server  struct {
		Host string `wconf:"HOST" default:"0.0.0.0"`
		Port int    `wconf:"PORT" default:"8080"`
	} `wconf:"SERVER"`
	Log struct {
		Level string `wconf:"LEVEL" default:"info"`
	} `wconf:"LOG"`
}
