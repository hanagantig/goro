package commands

// Options control one run of Init or Update.
type Options struct {
	// SkipGoMod leaves out the `go mod init` and `go mod tidy` steps. They run the go tool and
	// need the network, so in-process tests set it.
	SkipGoMod bool
}
