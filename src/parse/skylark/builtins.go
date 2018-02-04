package skylark

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"

	"core"
)

// registerBuiltins registers the various global builtins.
func registerBuiltins(config *core.Configuration) {
	skylark.Universe["CONFIG"] = makeConfig(config)
	skylark.Universe["build_rule"] = skylark.NewBuiltin("build_rule", buildRule)
	skylark.Universe["package"] = skylark.NewBuiltin("package", pkg)
	skylark.Universe["subinclude"] = skylark.NewBuiltin("subinclude", subinclude)
	skylark.Universe["fail"] = skylark.NewBuiltin("fail", fail)
	skylark.Universe["glob"] = skylark.NewBuiltin("glob", glob)
	skylark.Universe["get_labels"] = skylark.NewBuiltin("get_labels", getLabels)
	skylark.Universe["get_command"] = skylark.NewBuiltin("get_command", getCommand)
	skylark.Universe["set_command"] = skylark.NewBuiltin("set_command", setCommand)
	skylark.Universe["add_out"] = skylark.NewBuiltin("add_out", setCommand)
	skylark.Universe["add_licence"] = skylark.NewBuiltin("add_licence", setCommand)
	skylark.Universe["add_dep"] = skylark.NewBuiltin("add_dep", setCommand)
	skylark.Universe["add_exported_dep"] = skylark.NewBuiltin("add_dep", setCommand)
	skylark.Universe["log"] = skylarkstruct.FromStringDict(skylarkstruct.Default, skylark.StringDict{
		"debug":   logBuiltin("log.debug", log.Debug),
		"info":    logBuiltin("log.info", log.Info),
		"notice":  logBuiltin("log.notice", log.Notice),
		"warning": logBuiltin("log.warning", log.Warning),
		"error":   logBuiltin("log.error", log.Error),
		"fatal":   logBuiltin("log.fatal", log.Fatalf),
	})
}

// logBuiltin creates a builtin for logging things.
func logBuiltin(name string, f func(format string, args ...interface{})) skylark.Value {
	return skylark.NewBuiltin(name, func(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
		f("log not implemented")
		return skylark.None, nil
	})
}

// makeConfig creates a new config object from the given configuration.
func makeConfig(config *core.Configuration) skylark.Value {
	c := make(skylark.StringDict, 100)
	v := reflect.ValueOf(config).Elem()
	for i := 0; i < v.NumField(); i++ {
		if field := v.Field(i); field.Kind() == reflect.Struct {
			for j := 0; j < field.NumField(); j++ {
				if tag := field.Type().Field(j).Tag.Get("var"); tag != "" {
					subfield := field.Field(j)
					switch subfield.Kind() {
					case reflect.String:
						c[tag] = skylark.String(subfield.String())
					case reflect.Bool:
						c[tag] = skylark.Bool(subfield.Bool())
					case reflect.Slice:
						l := make([]skylark.Value, subfield.Len())
						for i := 0; i < subfield.Len(); i++ {
							l[i] = skylark.String(subfield.Index(i).String())
						}
						c[tag] = skylark.NewList(l)
					}
				}
			}
		}
	}
	// Arbitrary build config stuff
	for k, v := range config.BuildConfig {
		c[strings.Replace(strings.ToUpper(k), "-", "_", -1)] = skylark.String(v)
	}
	// Settings specific to package() which aren't in the config, but it's easier to
	// just put them in now.
	c["DEFAULT_VISIBILITY"] = skylark.None
	c["DEFAULT_TESTONLY"] = skylark.False
	c["DEFAULT_LICENCES"] = skylark.None
	// These can't be changed (although really you shouldn't be able to find out the OS at parse time)
	c["OS"] = skylark.String(runtime.GOOS)
	c["ARCH"] = skylark.String(runtime.GOARCH)
	return skylarkstruct.FromStringDict(skylarkstruct.Default, c)
}

// fail implements the only error-handling primitive you'll ever need.
func fail(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("msg is a required argument to fail()")
	}
	return nil, fmt.Errorf("%s", args[1])
}

// buildRule is the builtin that creates & registers a new build rule.
func buildRule(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	pkg := thread.Local("_pkg").(*core.Package)
	log.Info("adding target %s %s %s", pkg.Name, args, kwargs)
	return skylark.None, nil
}

// pkg implements the package() builtin.
func pkg(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	//log.Fatalf("package() not implemented")
	return skylark.None, nil
}

// subinclude implements the subinclude() builtin.
func subinclude(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	//log.Fatalf("subinclude() not implemented")
	return skylark.None, nil
}

// getLabels implements the get_labels builtin.
func getLabels(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return &skylark.List{}, nil
}

// getCommand implements the get_command builtin.
func getCommand(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return skylark.String(""), nil
}

// setCommand implements the get_command builtin.
func setCommand(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	return skylark.None, nil
}

// glob implements the glob() builtin.
func glob(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	var includes, exclude, excludes *skylark.List
	hidden := false
	if err := skylark.UnpackArgs("glob", args, kwargs, "includes", includes, "exclude?", exclude, "excludes?", excludes, "hidden?", &hidden); err != nil {
		return skylark.None, err
	}
	// TODO(peterebden): get rid of excludes...
	e := append(toStringList(exclude), toStringList(excludes)...)
	pkg := thread.Local("_pkg").(*core.Package)
	l := core.Glob(pkg.Name, toStringList(includes), e, e, hidden)
	ret := make([]skylark.Value, len(l))
	for i, f := range l {
		ret[i] = skylark.String(f)
	}
	return skylark.NewList(ret), nil
}

func toStringList(l *skylark.List) []string {
	if l == nil {
		return nil
	}
	ret := make([]string, l.Len())
	for i := range ret {
		ret[i] = string(l.Index(i).(skylark.String))
	}
	return ret
}
