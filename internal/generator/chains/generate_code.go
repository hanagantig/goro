package chains

import (
	"bytes"
	"go/format"
	entity "goro/internal/entity"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type generateCodeChain struct {
	data entity.AppData
}

func NewGenerateCodeChain(data entity.AppData) *generateCodeChain {
	return &generateCodeChain{
		data: data,
	}
}

func (g *generateCodeChain) Apply() error {
	err := filepath.WalkDir("/Users/hanagantig/tmp/gorotest",
		func(path string, d os.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if d.Name() == ".DS_Store" {
				return nil
			}

			buf := bytes.NewBuffer(nil)
			tmpl := template.Must(template.ParseFiles(path))

			g.data.Databases = []entity.Database{
				{
					Type:       "mysqlx",
					Name:       "mainDB",
					Connection: "*sqlx.DB",
				},
				{
					Type:       "mysql",
					Name:       "folbackDB",
					Connection: "*sql.DB",
				},
			}

			g.data.Dependencies = map[entity.DependencyName]entity.Dependency{
				"myRepo": entity.Dependency{
					Pkg:       "internal/adapter/repository",
					BuildFunc: "repository.NewMyRepo",
					Type:      "*repository.MyRepo",
					Deps:      []entity.DependencyName{"mainDB"},
				},
			}
			g.data.Module = g.data.Name

			err = tmpl.Execute(buf, g.data)
			if err != nil {
				log.Fatalf("Unable to parse data into template: %v\n", err)
			}

			formatted := buf.Bytes()
			if strings.Contains(d.Name(), ".go") {
				formatted, err = format.Source(buf.Bytes())
				if err != nil {
					log.Fatalf("Could not format processed template in file %s: %v\n", path, err)
				}
			}

			err = ioutil.WriteFile(path, formatted, 0644)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return err
}

func (g *generateCodeChain) Name() string {
	return "Generate code"
}

func (g *generateCodeChain) Rollback() error {
	return nil
}
