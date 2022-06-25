package chains

import (
	"bytes"
	entity "github.com/hanagantig/goro/internal/config"
	"github.com/hanagantig/goro/internal/generator"
	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type generateCodeChain struct{}

func NewGenerateCodeChain() *generateCodeChain {
	return &generateCodeChain{}
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

			fMap := generator.FunkMap
			fMap["toCamelCase"] = strcase.ToCamel

			buf := bytes.NewBuffer(nil)
			t := template.New(f.Name()).Funcs(fMap)
			content, err := fs.ReadFile(path)
			if err != nil {
				return err
			}
			tmpl := template.Must(t.Parse(string(content)))

			err = tmpl.Execute(buf, data)
			if err != nil {
				return err
			}

			formatted := buf.Bytes()
			if strings.Contains(f.Name(), ".go") {
				formatted, err = format.Source(buf.Bytes())
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
