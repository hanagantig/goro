package generator

import "embed"

//go:embed templates/app
var singleServiceTpl embed.FS

var singleServiceSkeleton = skeleton{
	root:     "templates/app",
	template: singleServiceTpl,
}

//var skeletons = map[string]skeleton{
//	"default": singleServiceSkeleton,
//}

type skeleton struct {
	root     string
	template embed.FS
}
