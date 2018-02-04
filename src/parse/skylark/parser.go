// Package skylark implements a parser based on Google's Skylark
// (see github.com/google/skylark).
// Currently this is experimental while I work out how closely it matches
// our language.
package skylark

import (
	"sort"
	"sync"

	"github.com/google/skylark"
	"github.com/google/skylark/resolve"
	"gopkg.in/op/go-logging.v1"

	"core"
)

var log = logging.MustGetLogger("skylark")

// A Parser is our wrapper around the Skylark parser.
type Parser struct {
	globals skylark.StringDict
	mutex   sync.Mutex
}

func NewParser(state *core.BuildState) *Parser {
	// We require these Skylark settings
	resolve.AllowNestedDef = true
	resolve.AllowLambda = true
	p := &Parser{
		globals: skylark.StringDict{},
	}
	t := p.newThread("builtins")
	// Preload builtins
	p.globals["CONFIG"] = makeConfig(state.Config)
	// Load all the globals now
	dir, _ := AssetDir("")
	sort.Strings(dir)
	// Temp hack - since c_library calls through to cc_library etc, we must parse them in order since
	// Skylark binds them on initial parse (unlike Python, which evaluates them at execution time).
	dir[1], dir[2] = dir[2], dir[1]
	for _, filename := range dir {
		if err := skylark.ExecFile(t, filename, MustAsset(filename), p.globals); err != nil {
			log.Fatalf("Failed to load builtin rules from %s: %s", filename, err)
		}
	}
	return p
}

// ParseFile parses a single BUILD file.
func (p *Parser) ParseFile(state *core.BuildState, pkg *core.Package, filename string) error {
	t := p.newThread(pkg.Name)
	return skylark.ExecFile(t, filename, nil, p.globals)
}

// print implements Skylark's print() builtin.
func (p *Parser) print(thread *skylark.Thread, msg string) {
	log.Info(msg)
}

// load implements Skylark's module loading.
func (p *Parser) load(thread *skylark.Thread, module string) (skylark.StringDict, error) {
	log.Fatalf("load not implemented: %s", module)
	return skylark.StringDict{}, nil
}

// newThread creates a new Skylark thread.
func (p *Parser) newThread(packageName string) *skylark.Thread {
	t := &skylark.Thread{
		Print: p.print,
		Load:  p.load,
	}
	t.SetLocal("PACKAGE_NAME", packageName)
	return t
}
