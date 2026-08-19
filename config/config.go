// Package config provides typed configuration loading from YAML files with
// environment-variable overlay for the Runvil ecosystem.
//
// Precedence is strictly: built-in zero values < file < environment.
// Later sources win.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultPrefix is prepended to implicit environment keys derived from field
// names. Fields tagged `env:"KEY"` are always read from the exact key.
const DefaultPrefix = "APP"

// Option configures loading behavior.
type Option func(*options)

type options struct {
	prefix string
}

// WithPrefix overrides DefaultPrefix used for implicit environment keys.
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// Load reads the YAML file at path and unmarshals it into dst (a non-nil
// pointer to a struct), binding values by `yaml` tag or lowercased field name.
// A missing file leaves dst at its zero value and returns nil; a malformed
// file returns a wrapped error naming the path.
func Load(path string, dst any) error {
	if err := checkTarget(path, dst); err != nil {
		return err
	}
	if path == "" {
		return fmt.Errorf("config: load: path must not be empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config: read %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("config: parse %q: %w", path, err)
	}
	return nil
}

// LoadOrDefault behaves like Load and is provided for tooling that tolerates
// an absent configuration file: a missing file yields the struct's zero value
// and never an error.
func LoadOrDefault(path string, dst any) error {
	return Load(path, dst)
}

// Override applies environment values on top of an already-loaded struct.
// A field tagged `env:"KEY"` is read from the exact key; an untagged field
// derives an implicit key from its yaml tag/name uppercased with `-` and `.`
// replaced by `_`, prefixed by DefaultPrefix (or the prefix from WithPrefix),
// joining nested struct segments with `_` (e.g. `Server.Addr` -> APP_SERVER_ADDR).
// Empty environment values are treated as unset. Slices and maps are not
// overlaid in this phase.
func Override(dst any, lookup func(string) (string, bool), opts ...Option) error {
	if dst == nil {
		return errors.New("config: override: dst must be non-nil")
	}
	if lookup == nil {
		return errors.New("config: override: lookup must be non-nil")
	}
	o := &options{prefix: DefaultPrefix}
	for _, opt := range opts {
		opt(o)
	}
	rv := reflect.ValueOf(dst)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errors.New("config: override: dst must be a non-nil pointer to a struct")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("config: override: dst must be a pointer to a struct")
	}
	return overrideFields(rv, "", "", o.prefix, lookup)
}

func checkTarget(path string, dst any) error {
	if dst == nil {
		return fmt.Errorf("config: load %q: dst must be non-nil", path)
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("config: load %q: dst must be a non-nil pointer", path)
	}
	if rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: load %q: dst must point to a struct", path)
	}
	return nil
}

func overrideFields(rv reflect.Value, baseKey, path, prefix string, lookup func(string) (string, bool)) error {
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := rv.Field(i)
		name := fieldName(sf)
		fullPath := name
		if path != "" {
			fullPath = path + "." + name
		}

		key := envKey(sf, baseKey, prefix)
		val, ok := "", false
		if key != "" {
			val, ok = lookup(key)
		}

		if isStructKind(fv) {
			child := derefPtr(fv)
			if child.IsValid() {
				if err := overrideFields(child, key, fullPath, prefix, lookup); err != nil {
					return err
				}
			}
			continue
		}

		if !ok || val == "" {
			continue
		}
		if fv.Kind() == reflect.Slice || fv.Kind() == reflect.Map {
			continue
		}
		if err := setScalar(fv, val); err != nil {
			return fmt.Errorf("config: field %s (env %s): %w", fullPath, key, err)
		}
	}
	return nil
}

// fieldName resolves the yaml tag name or the lowercased field name.
func fieldName(sf reflect.StructField) string {
	if tag, ok := sf.Tag.Lookup("yaml"); ok {
		if name := strings.Split(tag, ",")[0]; name != "" {
			return name
		}
	}
	return strings.ToLower(sf.Name)
}

// envKey returns the environment key for a field: the exact `env` tag when
// present, otherwise the implicit key derived from prefix + path segments.
func envKey(sf reflect.StructField, baseKey, prefix string) string {
	if tag, ok := sf.Tag.Lookup("env"); ok && tag != "" {
		return tag
	}
	return implicitKey(baseKey, fieldName(sf), prefix)
}

// implicitKey builds `PREFIX_SEGMENT` from a base key and a field name;
// nested segments are joined with `_` (e.g. `Server.Addr` -> `APP_SERVER_ADDR`).
func implicitKey(baseKey, name, prefix string) string {
	upper := strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToUpper(name))
	if baseKey != "" {
		return baseKey + "_" + upper
	}
	return prefix + "_" + upper
}

func isStructKind(v reflect.Value) bool {
	k := v.Kind()
	if k == reflect.Ptr {
		return isStructKind(derefPtr(v))
	}
	return k == reflect.Struct
}

func derefPtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func setScalar(fv reflect.Value, val string) error {
	target := reflect.New(fv.Type())
	if err := yaml.Unmarshal([]byte(val), target.Interface()); err != nil {
		return err
	}
	fv.Set(target.Elem())
	return nil
}
