package config

import (
	"strings"

	"github.com/iancoleman/strcase"
)

// DepFilePrefix starts the name of every dependency interface file, dep_<snake name>.go.
const DepFilePrefix = "dep_"

// MethodFile is the name of the file goro creates for a method of the use case, a service or an
// adapter: SignIn gets sign_in.go.
func MethodFile(method string) string {
	return strcase.ToSnake(method) + ".go"
}

// DepFile is the name of the file that holds the interface for dependency dep in the package of
// the use case or a service: UserRepo gets dep_user_repo.go.
func DepFile(dep string) string {
	return DepFilePrefix + strcase.ToSnake(dep) + ".go"
}

// buildConstraint says when Go builds a file named file, judged by its name the way go build
// does: "in tests" for a _test file, "on linux" for a name that ends in _linux, and the same for
// every other GOOS and GOARCH. It returns "" for a file that every build compiles.
func buildConstraint(file string) string {
	name := strings.TrimSuffix(file, ".go")
	i := strings.Index(name, "_")
	if i < 0 {
		return ""
	}

	parts := strings.Split(name[i+1:], "_")
	last := parts[len(parts)-1]
	switch {
	case last == "test":
		return "in tests"
	case knownOS[last], knownArch[last]:
		return "on " + last
	}

	return ""
}

// knownOS and knownArch are the GOOS and GOARCH values that go build reads from file names, as
// go/build lists them.
var knownOS = setOf("aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "js",
	"linux", "nacl", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "zos")

var knownArch = setOf("386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "loong64", "mips",
	"mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "ppc64", "ppc64le", "riscv",
	"riscv64", "s390", "s390x", "sparc", "sparc64", "wasm")

func setOf(values ...string) map[string]bool {
	res := make(map[string]bool, len(values))
	for _, v := range values {
		res[v] = true
	}

	return res
}
