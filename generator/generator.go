package generator

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/wkqco33/cli_template/templates"
)

// Options 프로젝트 생성 옵션
type Options struct {
	Template      string
	SQLite        bool
	Profile       bool
	ProfileWriter io.Writer // 프로파일 출력 대상 (비어 있으면 os.Stderr)
	ModuleName    string    // Go 모듈 경로 (비어 있으면 projectName 사용)
	Force         bool      // 대상 디렉토리가 이미 존재해도 덮어쓰기
	OutputDir     string    // 생성 위치 (비어 있으면 현재 디렉토리)
	DryRun        bool      // 실제 생성 없이 파일 목록만 출력
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

// sqliteSupportedTemplates --sqlite 옵션이 실제로 반영되는 템플릿 목록
var sqliteSupportedTemplates = map[string]bool{
	"minimal": true,
	"full":    true,
	"gin":     true,
	"fiber":   true,
	"echo":    true,
}

func init() {
	for _, meta := range templateCatalog {
		templateCatalogByName[meta.Name] = meta
	}
}

// SQLiteSupported 템플릿이 --sqlite 옵션을 지원하는지 여부를 반환한다
func SQLiteSupported(template string) bool {
	return sqliteSupportedTemplates[template]
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
	out         io.Writer
	startedAt   time.Time
	rendering   time.Duration
	postprocess time.Duration
}

func newGenerationProfile(enabled bool, out io.Writer) generationProfile {
	if out == nil {
		out = os.Stderr
	}
	return generationProfile{
		enabled:   enabled,
		out:       out,
		startedAt: time.Now(),
	}
}

func (p *generationProfile) print() {
	if !p.enabled {
		return
	}
	total := time.Since(p.startedAt)
	fmt.Fprintf(
		p.out,
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

// Resolved 검증을 통과한 생성 대상 정보
type Resolved struct {
	ModuleName string
	TargetPath string
}

// Resolve 옵션을 검증하고 모듈 이름과 대상 경로를 계산한다.
func Resolve(projectName string, opts Options) (Resolved, error) {
	moduleName := opts.ModuleName
	if moduleName == "" {
		moduleName = projectName
	}
	if err := ValidateProjectAndModuleName(projectName, moduleName); err != nil {
		return Resolved{}, err
	}
	if _, ok := templateMeta(opts.Template); !ok {
		return Resolved{}, NewUsageError("알 수 없는 템플릿: %s (사용 가능: %s)", opts.Template, TemplateNamesCSV())
	}
	targetPath := projectName
	if opts.OutputDir != "" {
		targetPath = filepath.Join(opts.OutputDir, projectName)
	}
	return Resolved{ModuleName: moduleName, TargetPath: targetPath}, nil
}

// Generate 지정한 이름과 옵션으로 프로젝트를 생성한다
func Generate(projectName string, opts Options) (retErr error) {
	profile := newGenerationProfile(opts.Profile, opts.ProfileWriter)
	defer func() {
		profile.print()
	}()

	resolved, err := Resolve(projectName, opts)
	if err != nil {
		return err
	}
	targetPath := resolved.TargetPath

	targetExists := false
	if _, err := os.Stat(targetPath); err == nil {
		if !opts.Force {
			return &ConflictError{Path: targetPath}
		}
		targetExists = true
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("대상 경로 확인 실패 (%s): %w", targetPath, err)
	}

	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("출력 디렉토리 생성 실패 (%s): %w", parentDir, err)
	}
	baseName := filepath.Base(targetPath)
	tempDir, err := os.MkdirTemp(parentDir, "."+baseName+".tmp-*")
	if err != nil {
		return fmt.Errorf("임시 작업 디렉토리 생성 실패 (parent=%s): %w", parentDir, err)
	}
	installed := false
	defer func() {
		if installed {
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
		ModuleName:  resolved.ModuleName,
		SQLite:      opts.SQLite,
	}

	renderStart := time.Now()
	if err := renderTemplatesFunc(tempDir, opts.Template, data); err != nil {
		return fmt.Errorf("템플릿 렌더링 단계 실패: %w", err)
	}
	profile.rendering = time.Since(renderStart)

	// 기존 디렉토리는 렌더링이 모두 끝난 뒤에만 백업으로 옮긴다.
	// 실패하면 백업을 되돌려 사용자 데이터를 보존한다.
	postprocessStart := time.Now()
	backupPath := ""
	if targetExists {
		backupPath, err = reserveBackupPath(parentDir, baseName)
		if err != nil {
			return err
		}
		if err := os.Rename(targetPath, backupPath); err != nil {
			return fmt.Errorf("기존 디렉토리 백업 실패 (%s -> %s): %w", targetPath, backupPath, err)
		}
	}

	if err := os.Rename(tempDir, targetPath); err != nil {
		err = fmt.Errorf("최종 경로 적용 실패 (%s -> %s): %w", tempDir, targetPath, err)
		if backupPath == "" {
			return err
		}
		if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
			return fmt.Errorf("%w; 기존 디렉토리 복원 실패 (백업 보존: %s): %v", err, backupPath, restoreErr)
		}
		return err
	}
	installed = true

	if backupPath != "" {
		if err := os.RemoveAll(backupPath); err != nil {
			return fmt.Errorf("백업 디렉토리 정리 실패 (백업 보존: %s): %w", backupPath, err)
		}
	}
	profile.postprocess = time.Since(postprocessStart)

	return nil
}

// reserveBackupPath 기존 디렉토리를 잠시 옮겨 둘 고유한 경로를 확보한다.
func reserveBackupPath(parentDir, baseName string) (string, error) {
	path, err := os.MkdirTemp(parentDir, "."+baseName+".bak-*")
	if err != nil {
		return "", fmt.Errorf("백업 디렉토리 생성 실패 (parent=%s): %w", parentDir, err)
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("백업 경로 준비 실패 (%s): %w", path, err)
	}
	return path, nil
}

// DryRun 실제 생성 없이 생성될 파일 목록을 반환한다.
// 파일시스템을 변경하지 않고 템플릿 실행만 검증한다.
func DryRun(projectName string, opts Options) ([]string, error) {
	resolved, err := Resolve(projectName, opts)
	if err != nil {
		return nil, err
	}

	data := TemplateData{
		ProjectName: projectName,
		ModuleName:  resolved.ModuleName,
		SQLite:      opts.SQLite,
	}

	files, err := planFiles(opts.Template, data)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		if err := renderFileTo(io.Discard, f.src, data); err != nil {
			return nil, fmt.Errorf("템플릿 렌더링 단계 실패: %w", err)
		}
		paths = append(paths, f.rel)
	}
	return paths, nil
}

// collectFiles 디렉토리 아래의 모든 파일을 상대 경로(슬래시 구분)로 수집한다
func collectFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// InitGit 지정한 디렉토리에 git 저장소를 초기화한다
func InitGit(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ExternalError{Tool: "git init", Err: fmt.Errorf("%s: %w\n%s", dir, err, string(out))}
	}
	return nil
}

// plannedFile 템플릿에서 만들어질 파일 하나. src는 embed.FS 경로, rel은 대상 상대 경로다.
type plannedFile struct {
	src string
	rel string
}

// planFiles 템플릿이 만들어낼 파일 목록을 계산한다. 파일시스템은 건드리지 않는다.
// 실제 생성과 dry-run이 같은 계획을 공유하므로 두 경로가 드리프트할 수 없다.
func planFiles(tmplName string, data TemplateData) ([]plannedFile, error) {
	var files []plannedFile

	err := fs.WalkDir(templates.FS, tmplName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, tmplName+"/")

		// SQLite 옵션 없으면 database/ 디렉토리 전체 건너뜀
		if !data.SQLite && (relPath == "database" || strings.HasPrefix(relPath, "database/")) {
			return nil
		}

		// embed.FS는 '.'로 시작하는 파일을 제외하므로 gitignore.tmpl을 .gitignore로 매핑한다
		if relPath == "gitignore.tmpl" {
			relPath = ".gitignore"
		}

		files = append(files, plannedFile{
			src: path,
			rel: strings.TrimSuffix(relPath, ".tmpl"),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("템플릿 파일 목록 계산 실패 (%s): %w", tmplName, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })
	return files, nil
}

func renderTemplates(destRoot, tmplName string, data TemplateData) error {
	files, err := planFiles(tmplName, data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destRoot, 0o755); err != nil {
		return fmt.Errorf("출력 디렉토리 생성 실패 (%s): %w", destRoot, err)
	}

	for _, f := range files {
		destPath := filepath.Join(destRoot, filepath.FromSlash(f.rel))
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("디렉토리 생성 실패 (%s): %w", filepath.Dir(destPath), err)
		}
		if err := renderFile(f.src, destPath, data); err != nil {
			return err
		}
	}
	return nil
}

func ValidateProjectAndModuleName(projectName, moduleName string) error {
	if err := validateSafeName("프로젝트 이름", projectName); err != nil {
		return err
	}
	if err := validateModuleName(moduleName); err != nil {
		return err
	}
	return nil
}

// validateModuleName Go 모듈 경로를 검증한다. 도메인 경로(github.com/user/repo)를 허용한다.
func validateModuleName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return NewUsageError("모듈 이름이(가) 비어 있습니다. %s", nameValidationGuide)
	}
	if name != trimmed {
		return NewUsageError("모듈 이름에 앞뒤 공백이 포함되어 있습니다. %s", nameValidationGuide)
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return NewUsageError("모듈 이름에 공백 문자가 포함되어 있습니다. %s", nameValidationGuide)
	}
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return NewUsageError("모듈 이름은 /로 시작하거나 끝날 수 없습니다. %s", nameValidationGuide)
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == "" {
			return NewUsageError("모듈 이름에 빈 경로 세그먼트가 포함되어 있습니다. %s", nameValidationGuide)
		}
		if !safeNamePattern.MatchString(seg) {
			return NewUsageError("모듈 이름 형식이 올바르지 않습니다. %s", nameValidationGuide)
		}
	}
	return nil
}

func validateSafeName(field, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return NewUsageError("%s이(가) 비어 있습니다. %s", field, nameValidationGuide)
	}
	if name != trimmed {
		return NewUsageError("%s에 앞뒤 공백이 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if filepath.IsAbs(name) {
		return NewUsageError("%s이(가) 절대 경로 형식입니다. %s", field, nameValidationGuide)
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return NewUsageError("%s에 공백 문자가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\`) {
		return NewUsageError("%s에 경로 구분자(/ 또는 \\)가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.Contains(name, "..") {
		return NewUsageError("%s에 상대 경로 패턴(..)이 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if strings.ContainsAny(name, `<>:"|?*`) {
		return NewUsageError("%s에 예약 문자(< > : \" | ? *)가 포함되어 있습니다. %s", field, nameValidationGuide)
	}
	if !safeNamePattern.MatchString(name) {
		return NewUsageError("%s 형식이 올바르지 않습니다. %s", field, nameValidationGuide)
	}
	return nil
}

func renderFile(srcPath, destPath string, data TemplateData) error {
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}

	if err := renderFileTo(f, srcPath, data); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// renderFileTo 템플릿 한 개를 w로 렌더링한다. .tmpl이 아니면 원본을 그대로 쓴다.
func renderFileTo(w io.Writer, srcPath string, data TemplateData) error {
	if !strings.HasSuffix(srcPath, ".tmpl") {
		content, err := templates.FS.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if _, err := w.Write(content); err != nil {
			return fmt.Errorf("파일 쓰기 실패 (%s): %w", srcPath, err)
		}
		return nil
	}

	tmpl, err := cachedTemplate(srcPath)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("템플릿 실행 실패 (%s): %w", srcPath, err)
	}
	return nil
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
