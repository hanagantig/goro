package chains

import (
	"bytes"
	"github.com/iancoleman/strcase"
	"github.com/spf13/afero"
	"go/format"
	entity "goro/internal/config"
	"os"
	"strings"
	"text/template"
)

type generateCodeChain struct {
	data entity.Config
}

func NewGenerateCodeChain(data entity.Config) *generateCodeChain {
	return &generateCodeChain{
		data: data,
	}
}

func (g *generateCodeChain) Apply(fs *afero.Afero) (*afero.Afero, error) {
	err := fs.Walk(g.data.App.WorkDir,
		func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				return nil
			}

			if f.Name() == ".DS_Store" || strings.Contains(path, ".idea") {
				return nil
			}

			fMap := template.FuncMap{
				"toCamelCase": strcase.ToCamel,
			}

			buf := bytes.NewBuffer(nil)
			t := template.New(f.Name()).Funcs(fMap)
			tmpl := template.Must(t.ParseFiles(path))

			err = tmpl.Execute(buf, g.data)
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
