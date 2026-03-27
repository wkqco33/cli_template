package generator

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"cli_template/templates"
)

// TemplateData 템플릿 렌더링 시 주입되는 데이터
type TemplateData struct {
	ProjectName string
	ModuleName  string
}

var validTemplates = map[string]bool{
	"minimal": true,
	"full":    true,
}

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
}

// submodules 템플릿별로 필요한 서브모듈 목록
var submodules = map[string][]struct{ path, url string }{
	"minimal": {
		{"wcli", "http://github.com/wkqco33/wcli"},
	},
	"full": {
		{"wcli", "http://github.com/wkqco33/wcli"},
		{"wconf", "http://github.com/wkqco33/wconf"},
	},
}

// Generate 지정한 이름과 템플릿으로 프로젝트를 생성한다
func Generate(projectName, tmplName string) error {
	if !validTemplates[tmplName] {
		return fmt.Errorf("알 수 없는 템플릿: %s (사용 가능: minimal, full)", tmplName)
	}

	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("디렉토리가 이미 존재합니다: %s", projectName)
	}

	data := TemplateData{
		ProjectName: projectName,
		ModuleName:  projectName,
	}

	if err := renderTemplates(projectName, tmplName, data); err != nil {
		return err
	}

	return initSubmodules(projectName, tmplName)
}

func renderTemplates(projectName, tmplName string, data TemplateData) error {
	return fs.WalkDir(templates.FS, tmplName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, tmplName+"/")
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

// initSubmodules git init 후 필요한 서브모듈을 추가한다
func initSubmodules(projectName, tmplName string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git을 찾을 수 없습니다. git을 설치하거나 wcli/wconf를 수동으로 추가하세요")
	}

	run := func(args ...string) error {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = projectName
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("'%s' 실패: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
		}
		return nil
	}

	if err := run("git", "init"); err != nil {
		return err
	}

	for _, sub := range submodules[tmplName] {
		if err := run("git", "submodule", "add", sub.url, sub.path); err != nil {
			return fmt.Errorf("서브모듈 추가 실패 (%s): %w", sub.path, err)
		}
	}

	return nil
}

func renderFile(srcPath, destPath string, data TemplateData) error {
	content, err := templates.FS.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(srcPath, ".tmpl") {
		return os.WriteFile(destPath, content, 0644)
	}

	tmpl, err := template.New(filepath.Base(srcPath)).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("템플릿 파싱 실패 (%s): %w", srcPath, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
