package generator

import (
	"strings"
	"testing"
)

func TestValidateProjectAndModuleName_Valid(t *testing.T) {
	validNames := []string{
		"my-cli",
		"app_v2",
		"abc",
		"a1",
		"tool.name",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			if err := ValidateProjectAndModuleName(name, name); err != nil {
				t.Fatalf("expected valid name, got error: %v", err)
			}
		})
	}
}

func TestValidateProjectAndModuleName_InvalidProjectName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContain string
	}{
		{name: "empty", input: "", wantContain: "비어 있습니다"},
		{name: "space", input: "my app", wantContain: "공백 문자가 포함"},
		{name: "slash", input: "my/app", wantContain: "경로 구분자"},
		{name: "backslash", input: `my\app`, wantContain: "경로 구분자"},
		{name: "dotdot", input: "my..app", wantContain: "상대 경로 패턴"},
		{name: "reserved", input: "my:app", wantContain: "예약 문자"},
		{name: "uppercase", input: "MyApp", wantContain: "형식이 올바르지 않습니다"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProjectAndModuleName(tc.input, tc.input)
			if err == nil {
				t.Fatalf("expected error for input %q", tc.input)
			}
			if !strings.Contains(err.Error(), tc.wantContain) {
				t.Fatalf("expected error to contain %q, got %q", tc.wantContain, err.Error())
			}
			if !strings.Contains(err.Error(), "해결 방법:") {
				t.Fatalf("expected actionable guide in error, got %q", err.Error())
			}
		})
	}
}

func TestValidateProjectAndModuleName_InvalidModuleName(t *testing.T) {
	err := ValidateProjectAndModuleName("valid-name", "bad module")
	if err == nil {
		t.Fatal("expected module name validation error")
	}
	if !strings.Contains(err.Error(), "모듈 이름") {
		t.Fatalf("expected module name context, got %q", err.Error())
	}
}
