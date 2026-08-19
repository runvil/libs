// Package validate validates structs using rules declared in `validate`
// struct tags. It is framework-agnostic and stdlib-only, so the web and CLI
// layers can share one rule model.
package validate

import (
	"fmt"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// FieldError describes a single validation failure.
type FieldError struct {
	// Field is the struct path of the failing field (e.g. "User.Email").
	Field string
	// Value is the value that failed validation (nil when unavailable).
	Value any
	// Rule is the rule that reported the failure (e.g. "min").
	Rule string
}

// Error implements the error interface.
func (e FieldError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: rule %q failed", e.Field, e.Rule)
	}
	return fmt.Sprintf("rule %q failed", e.Rule)
}

// ValidationError aggregates one or more field failures.
type ValidationError struct {
	Fields []FieldError
}

// Error implements the error interface, listing every failure.
func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Error()
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// Is reports whether target is a *ValidationError, supporting errors.Is.
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// Struct validates v (a struct or pointer to a struct) against its `validate`
// tags. It returns nil when every rule passes, a *ValidationError collecting
// the first failure per field, or a wrapped error for malformed tags.
func Struct(v any) error {
	if v == nil {
		return &ValidationError{Fields: []FieldError{{Rule: "struct"}}}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Struct {
		return &ValidationError{Fields: []FieldError{{Rule: "struct"}}}
	}
	eff, nilPtr := effective(rv)
	if nilPtr || eff.Kind() != reflect.Struct {
		return &ValidationError{Fields: []FieldError{{Rule: "struct"}}}
	}
	var errs []FieldError
	if err := walk(eff, "", &errs); err != nil {
		return err
	}
	if len(errs) == 0 {
		return nil
	}
	return &ValidationError{Fields: errs}
}

// Field validates a single value against a rule set, as used for scalar
// checks. It returns nil when the value passes and a *ValidationError
// otherwise.
func Field(value any, tag string) error {
	if strings.TrimSpace(tag) == "" {
		return nil
	}
	var rv reflect.Value
	if value != nil {
		rv, _ = effective(reflect.ValueOf(value))
	}
	fe, err := applyRules("", rv, tag)
	if err != nil {
		return err
	}
	if fe != nil {
		return &ValidationError{Fields: []FieldError{*fe}}
	}
	return nil
}

// walk validates the exported fields of a struct value.
func walk(rv reflect.Value, path string, errs *[]FieldError) error {
	t := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := rv.Field(i)
		name := sf.Name
		if path != "" {
			name = path + "." + sf.Name
		}
		tag := sf.Tag.Get("validate")
		eff, nilPtr := effective(fv)
		kind := valueKind(fv)

		if kind == reflect.Struct {
			if nilPtr {
				if hasRule(tag, "omitempty") {
					continue
				}
				*errs = append(*errs, FieldError{Field: name, Rule: "required"})
				continue
			}
			if fe, err := applyRules(name, eff, tag); err != nil {
				return err
			} else if fe != nil {
				*errs = append(*errs, *fe)
			}
			if err := walk(eff, name, errs); err != nil {
				return err
			}
			continue
		}

		fe, err := applyRules(name, eff, tag)
		if err != nil {
			return err
		}
		if fe != nil {
			*errs = append(*errs, *fe)
		}
	}
	return nil
}

// applyRules evaluates every rule in tag, reporting the first failing rule.
// A value with `omitempty` that is zero skips all rules. Malformed rules are
// returned as errors, never panics.
func applyRules(field string, v reflect.Value, tag string) (*FieldError, error) {
	if strings.TrimSpace(tag) == "" {
		return nil, nil
	}
	if hasRule(tag, "omitempty") && isZero(v) {
		return nil, nil
	}
	for _, rule := range strings.Split(tag, ",") {
		rule = strings.TrimSpace(rule)
		if rule == "" || rule == "omitempty" {
			continue
		}
		name, arg := parseRule(rule)
		fail, err := evalRule(name, arg, v)
		if err != nil {
			return nil, fmt.Errorf("validate: field %s: %w", field, err)
		}
		if fail {
			return &FieldError{Field: field, Value: interfaceValue(v), Rule: name}, nil
		}
	}
	return nil, nil
}

func evalRule(name, arg string, v reflect.Value) (bool, error) {
	switch name {
	case "required":
		return isZero(v), nil
	case "min", "max":
		return evalBound(name == "min", arg, v)
	case "len":
		l, ok := lenOf(v)
		if !ok {
			return false, fmt.Errorf("rule %q requires a string, slice, map, or array", name)
		}
		want, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
		if err != nil {
			return false, fmt.Errorf("rule %q: %w", name, err)
		}
		return int64(l) != want, nil
	case "email":
		if v.Kind() != reflect.String {
			return false, fmt.Errorf("rule %q requires a string field", name)
		}
		_, err := mail.ParseAddress(v.String())
		return err != nil, nil
	case "pattern":
		if v.Kind() != reflect.String {
			return false, fmt.Errorf("rule %q requires a string field", name)
		}
		re, err := regexp.Compile(`\A(?:` + arg + `)\z`)
		if err != nil {
			return false, fmt.Errorf("rule %q: %w", name, err)
		}
		return !re.MatchString(v.String()), nil
	case "oneof":
		return evalOneof(arg, v)
	default:
		return false, fmt.Errorf("unknown rule %q", name)
	}
}

func evalBound(isMin bool, arg string, v reflect.Value) (bool, error) {
	want, err := strconv.ParseFloat(strings.TrimSpace(arg), 64)
	if err != nil {
		return false, fmt.Errorf("invalid bound %q", arg)
	}
	var got float64
	switch v.Kind() {
	case reflect.String:
		got = float64(len(v.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		got = float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		got = float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		got = v.Float()
	default:
		return false, fmt.Errorf("min/max require a number or string, got %s", v.Kind())
	}
	if isMin {
		return got < want, nil
	}
	return got > want, nil
}

func evalOneof(arg string, v reflect.Value) (bool, error) {
	items := strings.Fields(arg)
	var s string
	switch v.Kind() {
	case reflect.String:
		s = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		s = strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		return false, fmt.Errorf("oneof requires a string or comparable number, got %s", v.Kind())
	}
	for _, it := range items {
		if s == it {
			return false, nil
		}
	}
	return true, nil
}

func lenOf(v reflect.Value) (int, bool) {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len(), true
	}
	return 0, false
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return v.Len() == 0
	}
	return v.IsZero()
}

func interfaceValue(v reflect.Value) any {
	if v.IsValid() && v.CanInterface() {
		return v.Interface()
	}
	return nil
}

func hasRule(tag, want string) bool {
	for _, r := range strings.Split(tag, ",") {
		if strings.TrimSpace(r) == want {
			return true
		}
	}
	return false
}

func parseRule(rule string) (name, arg string) {
	if i := strings.IndexByte(rule, '='); i >= 0 {
		return rule[:i], rule[i+1:]
	}
	return rule, ""
}

func effective(v reflect.Value) (reflect.Value, bool) {
	nilPtr := false
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			nilPtr = true
			break
		}
		v = v.Elem()
	}
	return v, nilPtr
}

func valueKind(v reflect.Value) reflect.Kind {
	t := v.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind()
}
