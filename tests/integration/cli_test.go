package integration

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/tiernacity/ratchet/internal/cli"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"ratchet": func() { os.Exit(cli.Main()) },
	})
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			// Set up any environment variables or test-specific configuration
			return nil
		},
	})
}