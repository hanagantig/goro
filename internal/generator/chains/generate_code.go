package chains

import (
	"bytes"
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/hanagantig/goro/pkg/afero"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type generateCodeChain struct{}

func NewGenerateCodeChain() *generateCodeChain {
	return &generateCodeChain{}
}

func generate(path string, content []byte, data entity.Config) ([]byte, error) {
	fMap := generator.FuncMap

	buf := bytes.NewBuffer(nil)
	t := template.New(path).Funcs(fMap)
	tmpl := template.Must(t.Parse(string(content)))

	err := tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *generateCodeChain) Apply(fs *afero.Afero, data entity.Config) (*afero.Afero, error) {
	err := fs.Walk("/",
		func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if f.IsDir() {
				return nil
			}

			if f.Name() == ".DS_Store" || strings.Contains(path, ".idea") {
				return nil
			}

			content, err := fs.ReadFile(path)

			formatted, err := generate(f.Name(), content, data)
			if err != nil {
				return err
			}

			if strings.Contains(f.Name(), ".go") {
				formatted, err = format.Source(formatted)
				if err != nil {
					return err
				}
			}

			err = fs.WriteFile(path, formatted, 0644)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return fs, err
}

func (g *generateCodeChain) Name() string {
	return "Generate code"
}

func (g *generateCodeChain) Rollback() error {
	return nil
}
