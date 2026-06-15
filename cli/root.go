// Package cli assembles the xiaoyuzhou command tree from the xiaoyuzhou
// domain on top of the any-cli/kit framework.
package cli

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/xiaoyuzhou-cli/xiaoyuzhou"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp assembles the kit application from the xiaoyuzhou domain. The
// domain's Register installs the client factory and every operation, so the
// binary and a host share one source of truth. kit.Run turns the App into
// the CLI, plus the serve and mcp surfaces and the typed-error-to-exit-code
// mapping.
//
// To add a command, declare it in xiaoyuzhou/domain.go with kit.Handle and it
// appears here automatically.
func NewApp() *kit.App {
	id := xiaoyuzhou.Domain{}.Info().Identity
	id.Version = Version

	app := kit.New(id)
	(xiaoyuzhou.Domain{}).Register(app)
	app.AddCommand(newVersionCmd())
	return app
}
