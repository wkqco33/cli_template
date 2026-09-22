package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const configDirName = "wtemp"
const configFileName = "config.yaml"

func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("사용자 설정 디렉터리 확인 실패: %w", err)
	}
	return filepath.Join(base, configDirName, configFileName), nil
}

func ResolvePath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if env := os.Getenv("WTEMP_CONFIG"); env != "" {
		return env, nil
	}
	return DefaultPath()
}
