package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"cli_template/templates"
)

// Options 프로젝트 생성 옵션
type Options struct {
	Template string
	SQLite   bool
	Profile  bool
}

// TemplateData 템플릿 렌더링 시 주입되는 데이터
type TemplateData struct {
	ProjectName string
	ModuleName  string
	SQLite      bool
}

// TemplateMeta 템플릿 카탈로그 메타데이터
type TemplateMeta struct {
	Name string
	Desc string
}

var templateCatalog = []TemplateMeta{
	{
		Name: "minimal",
		Desc: "루트 커맨드만 있는 최소 구조",
	},
	{
		Name: "full",
		Desc: "서브커맨드 + wcli 설정이 포함된 전체 구조",
	},
	{
		Name: "gin",
		Desc: "CLI + Gin 웹서버 (REST API 스켈레톤)",
	},
	{
		Name: "fiber",
		Desc: "CLI + Fiber 웹서버 (REST API 스켈레톤)",
	},
	{
		Name: "echo",
		Desc: "CLI + Echo 웹서버 (REST API 스켈레톤)",
	},
	{
		Name: "fyne",
		Desc: "CLI + Fyne GUI 앱",
	},
	{
		Name: "library",
		Desc: "Go 라이브러리 스켈레톤",
	},
}

var templateCatalogByName = map[string]TemplateMeta{}

func init() {
	for _, meta := range templateCatalog {
		templateCatalogByName[meta.Name] = meta
	}
}

var funcMap = template.FuncMap{
	"upper":   strings.ToUpper,
	"pkgname": func(s string) string { return strings.ReplaceAll(s, "-", "_") },
}

var safeNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)

const nameValidationGuide = "해결 방법: 영문 소문자(a-z), 숫자(0-9), '-', '_', '.'만 사용하고 시작/끝은 영문 소문자 또는 숫자로 입력하세요. 예: my-cli, app_v2"

var (
	renderTemplatesFunc = renderTemplates
)

type generationProfile struct {
	enabled     bool
	startedAt   time.Time
	rendering   time.Duration
	postprocess time.Duration
}

func newGenerationProfile(enabled bool) generationProfile {
	return generationProfile{
		enabled:   enabled,
		startedAt: time.Now(),
	}
}

func (p *generationProfile) print() {
	if !p.enabled {
		return
	}
	total := time.Since(p.startedAt)
	fmt.Fprintf(
		os.Stderr,
		"[profile] total=%s render=%s postprocess=%s\n",
		total.Truncate(time.Microsecond),
		p.rendering.Truncate(time.Microsecond),
		p.postprocess.Truncate(time.Microsecond),
	)
}

var parsedTemplateCache sync.Map // map[string]*template.Template

// Templates 현재 지원하는 템플릿 메타데이터 목록을 반환한다
func Templates() []TemplateMeta {
	out := make([]TemplateMeta, 0, len(templateCatalog))
	for _, meta := range templateCatalog {
		out = append(out, meta)
	}
	return out
}

// TemplateNamesCSV 템플릿 이름 목록을 쉼표 문자열로 반환한다
func TemplateNamesCSV() string {
	names := make([]string, 0, len(templateCatalog))
	for _, meta := range templateCatalog {
		names = append(names, meta.Name)
	}
	return strings.Join(names, ", ")
}

func templateMeta(name string) (TemplateMeta, bool) {
	meta, ok := templateCatalogByName[name]
	return meta, ok
}

// Generate 지정한 이름과 옵션으로 프로젝트를 생성한다
func Generate(projectName string, opts Options) (retErr error) {
	profile := newGenerationProfile(opts.Profile)
	defer func() {
		profile.print()
	}()

	if _, ok := templateMeta(opts.Template); !ok {
		return fmt.Errorf("알 수 없는 템플릿: %s (사용 가능: %s)", opts.Template, TemplateNamesCSV())
	}

	if err := ValidateProjectAndModuleName(projectName, projectName); err != nil {
		return err
	}

	targetPath := projectName
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("디렉토리가 이미 존재합니다: %s", targetPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("대상 경로 확인 실패 (%s): %w", targetPath, err)
	}

	parentDir := filepath.Dir(targetPath)
	baseName := filepath.Base(targetPath)
	tempDir, err := os.MkdirTemp(parentDir, "."+baseName+".tmp-*")
	if err != nil {
		return fmt.Errorf("임시 작업 디렉토리 생성 실패 (parent=%s): %w", parentDir, err)
	}
	renameDone := false
	defer func() {
		if renameDone {
			return
		}
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			if retErr != nil {
				retErr = fmt.Errorf("%w; 임시 디렉토리 정리 실패 (%s): %v", retErr, tempDir, cleanupErr)
				return
			}
			retErr = fmt.Errorf("임시 디렉토리 정리 실패 (%s): %w", tempDir, cleanupErr)
		}
	}()

	data := TemplateData{
		ProjectName: projectName,
		ModuleName:  projectName,
		SQLite:      opts.SQLite,
	}

	renderStart := time.Now()
	if err := renderTemplatesFunc(tempDir, opts.Template, data); err != nil {
		return fmt.Errorf("템플릿 렌더링 단계 실패: %w", err)
	}
	profile.rendering = time.Since(renderStart)

	postprocessStart := time.Now()
	if err := os.Rename(tempDir, targetPath); err != nil {
		return fmt.Errorf("최종 경로 적용 실패 (%s -> %s): %w", tempDir, targetPath, err)
	}
	profile.postprocess = time.Since(postprocessStart)
	renameDone = true

	return nil
}

func renderTemplates(projectName, tmplName string, data TemplateData) error {
	return fs.WalkDir(templates.FS, tmplName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, tmplName+"/")

		// SQLite 옵션 없으면 database/ 디렉토리 전체 건너뜀
		if !data.SQLite && (relPath == "database" || strings.HasPrefix(relPath, "database/")) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(projectName, strings.TrimSuffix(relPath, ".tmpl"))

		if d.IsDir() {
			if path == tmplName {
				return os.MkdirAll(projectName, 0755)
			}
			return os.MkdirAll(destPath, 0755)
		}

		return renderFile(path, destPath, data)
	})
}

func ValidateProjectAndModuleName(projectName, moduleName string) error {
	if err := validateSafeName("프로젝트 이름", projectName); err != nil {
		return err
	}
	if err := validateSafeName("모듈 이름", moduleName); err != nil {
		return err
	}
	return nil
}

func validateSafeName(field, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("%s이(가) 비어 있습니다. %s", field, nameValidationGuide)
	}
	if name != trimmed {
		return fmt.Errorf("%s에 앞뒤 공백이 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("%s이(가) 절대 경로 형식입니다. %s", field, nameValidationGuide)
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("%s에 공백 문자가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\`) {
		return fmt.Errorf("%s에 경로 구분자(/ 또는 \\)가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("%s에 상대 경로 패턴(..)이 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.ContainsAny(name, `<>:"|?*`) {
		return fmt.Errorf("%s에 예약 문자(< > : \" | ? *)가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if !safeNamePattern.MatchString(name) {
		return fmt.Errorf("%s 형식이 올바르지 않습니다. %s", field, nameValidationGuide)
	}
	return nil
}

func renderFile(srcPath, destPath string, data TemplateData) error {
	if !strings.HasSuffix(srcPath, ".tmpl") {
		content, err := templates.FS.ReadFile(srcPath)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, content, 0644)
	}

	tmpl, err := cachedTemplate(srcPath)
	if err != nil {
		return err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func cachedTemplate(srcPath string) (*template.Template, error) {
	if cached, ok := parsedTemplateCache.Load(srcPath); ok {
		return cached.(*template.Template), nil
	}

	content, err := templates.FS.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(filepath.Base(srcPath)).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("템플릿 파싱 실패 (%s): %w", srcPath, err)
	}
	actual, _ := parsedTemplateCache.LoadOrStore(srcPath, tmpl)
	return actual.(*template.Template), nil
}
