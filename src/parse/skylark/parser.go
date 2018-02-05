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
	registerBuiltins(state.Config)
	p := &Parser{}
	t := p.newThread(nil, nil)
	t.SetLocal("PACKAGE_NAME", "builtins")

	// The ordering here is deliberate; functions must be defined before any
	// other function that might call it, because Skylark binds them early
	// (in contrast to Python, which would bind them only when used).
	p.loadBuiltins(t, "builtins.sky")
	p.loadBuiltins(t, "misc_rules.build_defs")
	p.loadBuiltins(t, "cc_rules.build_defs")
	p.loadBuiltins(t, "c_rules.build_defs")
	p.loadBuiltins(t, "go_rules.build_defs")
	p.loadBuiltins(t, "python_rules.build_defs")
	p.loadBuiltins(t, "java_rules.build_defs")
	p.loadBuiltins(t, "sh_rules.build_defs")
	p.loadBuiltins(t, "proto_rules.build_defs")
	p.loadBuiltins(t, "bazel_compat.build_defs")
	for _, preload := range state.Config.Parse.PreloadBuildDefs {
		log.Debug("Preloading build defs from %s...", preload)
		p.loadGlobals(t, preload, nil)
	}
	return p
}

// ParseFile parses a single BUILD file.
func (p *Parser) ParseFile(state *core.BuildState, pkg *core.Package, filename string) error {
	t := p.newThread(state, pkg)
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
	t := p.newThread(getState(thread), getPkg(thread))
	globals := skylark.StringDict{}
	return globals, skylark.ExecFile(t, module, b, globals)
}

// loadBuiltins loads some builtins from a prepackaged asset and adds them to Skylark's universe.
func (p *Parser) loadBuiltins(thread *skylark.Thread, filename string) {
	p.loadGlobals(thread, filename, MustAsset(filename))
}

// loadGlobals loads some builtins from a file (or data) and adds them to Skylark's universe.
func (p *Parser) loadGlobals(thread *skylark.Thread, filename string, data interface{}) {
	globals := skylark.StringDict{}
	if err := skylark.ExecFile(thread, filename, data, globals); err != nil {
		log.Fatalf("Failed to load builtin rules: %s", err)
	}
	for k, v := range globals {
		skylark.Universe[k] = v
	}
}

// newThread creates a new Skylark thread.
func (p *Parser) newThread(state *core.BuildState, pkg *core.Package) *skylark.Thread {
	t := &skylark.Thread{
		Print: p.print,
		Load:  p.load,
	}
	if state != nil {
		t.SetLocal("_state", state)
		t.SetLocal("PACKAGE_NAME", pkg.Name)
		t.SetLocal("_pkg", pkg)
	}
	return t
}

// getState retrieves the build state object from a Skylark thread.
func getState(thread *skylark.Thread) *core.BuildState {
	return thread.Local("_state").(*core.BuildState)
}

// getPkg gets the package object from a Skylark thread.
func getPkg(thread *skylark.Thread) *core.Package {
	return thread.Local("_pkg").(*core.Package)
}
