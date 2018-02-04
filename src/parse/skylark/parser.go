// Package skylark implements a parser based on Google's Skylark
// (see github.com/google/skylark).
// Currently this is experimental while I work out how closely it matches
// our language.
package skylark

import (
	"github.com/google/skylark"
	"github.com/google/skylark/resolve"
	"gopkg.in/op/go-logging.v1"

	"core"
)

var log = logging.MustGetLogger("skylark")

// A Parser is our wrapper around the Skylark parser.
type Parser struct {
}

func NewParser(state *core.BuildState) *Parser {
	// We require these Skylark settings
	resolve.AllowNestedDef = true
	resolve.AllowLambda = true
	// Skylark global builtins
	skylark.Universe["build_rule"] = skylark.NewBuiltin("build_rule", buildRule)
	skylark.Universe["fail"] = skylark.NewBuiltin("fail", fail)
	skylark.Universe["CONFIG"] = makeConfig(state.Config)
	p := &Parser{}
	t := p.newThread(nil)
	t.SetLocal("PACKAGE_NAME", "builtins")
	// Load builtins.sky - that triggers loading of everything else.
	globals := skylark.StringDict{}
	if err := skylark.ExecFile(t, "builtins.sky", MustAsset("builtins.sky"), globals); err != nil {
		log.Fatalf("Failed to load builtin rules: %s", err)
	}
	return p
}

// ParseFile parses a single BUILD file.
func (p *Parser) ParseFile(state *core.BuildState, pkg *core.Package, filename string) error {
	t := p.newThread(pkg)
	globals := skylark.StringDict{}
	return skylark.ExecFile(t, filename, nil, globals)
}

// print implements Skylark's print() builtin.
func (p *Parser) print(thread *skylark.Thread, msg string) {
	log.Info(msg)
}

// load implements Skylark's module loading.
func (p *Parser) load(thread *skylark.Thread, module string) (skylark.StringDict, error) {
	b, err := Asset(module)
	if err != nil {
		return nil, err
	}
	// TODO(peterebden): Proper subinclude deferral
	t := p.newThread(nil)
	globals := skylark.StringDict{}
	return globals, skylark.ExecFile(t, module, b, globals)
}

// newThread creates a new Skylark thread.
func (p *Parser) newThread(pkg *core.Package) *skylark.Thread {
	t := &skylark.Thread{
		Print: p.print,
		Load:  p.load,
	}
	if pkg != nil {
		t.SetLocal("PACKAGE_NAME", pkg.Name)
		t.SetLocal("_pkg", pkg)
	}
	return t
}
