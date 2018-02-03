// Package skylark implements a parser based on Google's Skylark
// (see github.com/google/skylark).
// Currently this is experimental while I work out how closely it matches
// our language.
package skylark

import (
	"sort"
	"sync"

	"github.com/google/skylark"
	"gopkg.in/op/go-logging.v1"

	"core"
)

var log = logging.MustGetLogger("skylark")

// A Parser is our wrapper around the Skylark parser.
type Parser struct {
	threads     []*skylark.Thread
	busyThreads []*skylark.Thread
	globals     skylark.StringDict
	mutex       sync.Mutex
}

func NewParser(state *core.BuildState) *Parser {
	p := &Parser{
		threads: make([]*skylark.Thread, state.Config.Please.NumThreads),
	}
	for i := range p.threads {
		p.threads[i] = &skylark.Thread{
			Print: p.print,
			Load:  p.load,
		}
	}
	// Load all the globals now
	dir, _ := AssetDir("")
	sort.Strings(dir)
	// Temp hack - since c_library calls through to cc_library etc, we must parse them in order.
	dir[0], dir[1] = dir[1], dir[0]
	for _, filename := range dir {
		if err := skylark.ExecFile(p.threads[0], filename, MustAsset(filename), p.globals); err != nil {
			log.Fatalf("Failed to load builtin rules from %s: %s", filename, err)
		}
	}
	return p
}

// ParseFile parses a single BUILD file.
func (p *Parser) ParseFile(state *core.BuildState, pkg *core.Package, filename string) error {
	t := p.getThread()
	defer p.releaseThread(t)
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

// getThread grabs a Skylark thread.
// This is a bit awkward; we have a concept of a 'thread id' but it isn't passed through to this point
// so we have to do some pooling of them manually.
func (p *Parser) getThread() *skylark.Thread {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.threads) == 0 {
		// This shouldn't really happen, because we have as many threads as workers.
		log.Fatalf("No threads available")
	}
	idx := len(p.threads) - 1
	t := p.threads[idx]
	p.threads = p.threads[:idx]
	p.busyThreads = append(p.busyThreads, t)
	return t
}

// releaseThread releases a thread previously returned by getThread.
func (p *Parser) releaseThread(thread *skylark.Thread) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	threads := make([]*skylark.Thread, 0, len(p.busyThreads)-1)
	for _, t := range p.busyThreads {
		if t != thread {
			threads = append(threads, t)
		}
	}
	p.busyThreads = threads
	p.threads = append(p.threads, thread)
}
