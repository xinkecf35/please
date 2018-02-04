package skylark

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"

	"core"
)

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
