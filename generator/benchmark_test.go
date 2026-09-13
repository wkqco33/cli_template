package generator

import (
	"os"
	"testing"
)

// BenchmarkGenerate 프로젝트 생성 전체 경로(렌더링 + 원자적 교체)를 측정한다.
func BenchmarkGenerate(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(b.TempDir()); err != nil {
		b.Fatalf("chdir failed: %v", err)
	}
	b.Cleanup(func() { _ = os.Chdir(cwd) })

	for _, tmpl := range []string{"minimal", "full", "library"} {
		b.Run(tmpl, func(b *testing.B) {
			opts := Options{Template: tmpl, Force: true}
			b.ReportAllocs()
			for b.Loop() {
				if err := Generate("benchapp", opts); err != nil {
					b.Fatalf("generate failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPlanFiles 생성 계획 계산(임베드 FS 순회) 비용을 측정한다.
func BenchmarkPlanFiles(b *testing.B) {
	data := TemplateData{ProjectName: "app", ModuleName: "app", SQLite: true}

	for _, tmpl := range []string{"minimal", "full"} {
		b.Run(tmpl, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := planFiles(tmpl, data); err != nil {
					b.Fatalf("planFiles failed: %v", err)
				}
			}
		})
	}
}
