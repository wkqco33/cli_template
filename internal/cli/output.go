package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/wkqco33/cli_template/generator"
)

// 지원하는 출력 형식
const (
	FormatTable = "table"
	FormatPlain = "plain"
	FormatJSON  = "json"
)

// resolveFormat 명시된 형식을 검증한다. 지정하지 않으면 TTY면 table,
// 파이프·CI면 plain으로 자동 선택한다.
func resolveFormat(explicit string, w io.Writer) (string, error) {
	switch explicit {
	case FormatTable, FormatPlain, FormatJSON:
		return explicit, nil
	case "":
		if isTerminal(w) {
			return FormatTable, nil
		}
		return FormatPlain, nil
	default:
		return "", generator.NewUsageError(
			"알 수 없는 --format 값: %s (사용 가능: %s, %s, %s)",
			explicit, FormatTable, FormatPlain, FormatJSON,
		)
	}
}

// formatUsage --format 플래그 설명
func formatUsage() string {
	return fmt.Sprintf(
		"출력 형식 (%s, %s, %s). 기본값: TTY면 %s, 아니면 %s",
		FormatTable, FormatPlain, FormatJSON, FormatTable, FormatPlain,
	)
}

// writeJSON JSON 한 줄을 w에 출력한다.
func writeJSON(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}
