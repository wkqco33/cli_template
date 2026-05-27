package config

// Config 애플리케이션 설정 구조체
type Config struct {
	AppName string `wcli:"NAME" default:"super_cli"`
	Server  struct {
		Host string `wcli:"HOST" default:"0.0.0.0"`
		Port int    `wcli:"PORT" default:"8080"`
	} `wcli:"SERVER"`
	Log struct {
		Level string `wcli:"LEVEL" default:"info"`
	} `wcli:"LOG"`
}
