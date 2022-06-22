package generator

type Option func(generator *Generator)

//func WithSkeleton(name string) Option {
//	return func(g *Generator) {
//		g.skeleton = appTmplFs
//	}
//}
//
//func WithStorages(storages config.StorageList) Option {
//	return func(g *Generator) {
//		g.skeleton = appTmplFs
//	}
//}
