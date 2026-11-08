package tsoptions

import (
	"math"
	"reflect"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/collections"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/diagnostics"
)

func parseTypefnOptions(value any) (*core.TypefnOptions, []*ast.Diagnostic) {
	invalid := func(message string) (*core.TypefnOptions, []*ast.Diagnostic) {
		return nil, []*ast.Diagnostic{ast.NewCompilerDiagnostic(diagnostics.NewAdHocMessage("typefn: " + message))}
	}
	object, ok := typefnObject(value)
	if !ok || object == nil {
		return invalid("expected an object with version: 1")
	}
	version, exists := object.Get("version")
	if !exists || version != float64(1) {
		return invalid("only language version 1 is supported")
	}
	result := &core.TypefnOptions{Version: 1}
	for key, value := range object.Entries() {
		switch key {
		case "version":
		case "limits":
			limits, ok := typefnObject(value)
			if !ok || limits == nil {
				return invalid("limits must be an object")
			}
			fields := reflect.ValueOf(&result.Limits).Elem()
			for name, value := range limits.Entries() {
				index := -1
				for i := range fields.NumField() {
					if strings.Split(fields.Type().Field(i).Tag.Get("json"), ",")[0] == name {
						index = i
						break
					}
				}
				if index < 0 {
					return invalid("unknown limit " + name)
				}
				n, ok := value.(float64)
				if !ok || math.IsNaN(n) || math.IsInf(n, 0) || n < 1 || n > 9007199254740991 || math.Trunc(n) != n || n >= float64(math.MaxInt) {
					return invalid("limit " + name + " must be a positive safe integer representable on this host")
				}
				fields.Field(index).SetInt(int64(n))
			}
		default:
			return invalid("unknown option " + key)
		}
	}
	return result, nil
}

// Config parsing preserves insertion order; the raw JSON API uses ordinary maps.
// Accept both at the shared validation boundary without duplicating semantics.
func typefnObject(value any) (*collections.OrderedMap[string, any], bool) {
	switch object := value.(type) {
	case *collections.OrderedMap[string, any]:
		return object, object != nil
	case map[string]any:
		if object == nil {
			return nil, false
		}
		result := collections.NewOrderedMapWithSizeHint[string, any](len(object))
		for key, value := range object {
			result.Set(key, value)
		}
		return result, true
	}
	return nil, false
}
