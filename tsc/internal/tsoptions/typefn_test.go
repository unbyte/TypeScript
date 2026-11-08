package tsoptions

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/json"

	"github.com/microsoft/TypeScript/tsc/internal/collections"
)

func TestTypefnOptions(t *testing.T) {
	t.Parallel()
	object := func(values map[string]any) *collections.OrderedMap[string, any] {
		m := collections.NewOrderedMapWithSizeHint[string, any](len(values))
		for k, v := range values {
			m.Set(k, v)
		}
		return m
	}
	for _, test := range []struct {
		name  string
		input any
		valid bool
	}{
		{"version", object(map[string]any{"version": float64(1)}), true},
		{"limits", object(map[string]any{"version": float64(1), "limits": object(map[string]any{"syntaxDepth": float64(32)})}), true},
		{"missing", object(nil), false},
		{"unsupported", object(map[string]any{"version": float64(2)}), false},
		{"null", nil, false},
		{"unknown", object(map[string]any{"version": float64(1), "enable": true}), false},
		{"zero", object(map[string]any{"version": float64(1), "limits": object(map[string]any{"callDepth": float64(0)})}), false},
		{"fraction", object(map[string]any{"version": float64(1), "limits": object(map[string]any{"callDepth": 1.5})}), false},
		{"unsafe", object(map[string]any{"version": float64(1), "limits": object(map[string]any{"callDepth": float64(9007199254740992)})}), false},
		{"unknown-limit", object(map[string]any{"version": float64(1), "limits": object(map[string]any{"calls": float64(5)})}), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts, errors := parseTypefnOptions(test.input)
			if (len(errors) == 0) != test.valid || (opts != nil) != test.valid {
				t.Fatalf("options=%#v, errors=%v", opts, errors)
			}
		})
	}
}

func TestTypefnProfileInvalidation(t *testing.T) {
	t.Parallel()
	a := &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1, Limits: core.TypefnLimits{CallDepth: 16}}}
	b := &core.CompilerOptions{Typefn: &core.TypefnOptions{Version: 1, Limits: core.TypefnLimits{CallDepth: 16}}}
	if !a.Equals(b) || CompilerOptionsAffectSemanticDiagnostics(a, b) || CompilerOptionsAffectEmit(a, b) {
		t.Fatal("equivalent independently parsed profiles must compare equal")
	}
	b.Typefn.Limits.CallDepth++
	if a.Equals(b) || !CompilerOptionsAffectSemanticDiagnostics(a, b) || !CompilerOptionsAffectEmit(a, b) {
		t.Fatal("resource limit changes must invalidate checking and emit")
	}
	b.Typefn = nil
	if a.Equals(b) || !CompilerOptionsAffectSemanticDiagnostics(a, b) || !CompilerOptionsAffectEmit(a, b) {
		t.Fatal("disabling the profile must invalidate checking and emit")
	}
	found := false
	ForEachCompilerOptionAffectingBuildInfo(a, func(option *CommandLineOption, value any) {
		if option.Name == "typefn" {
			found = true
		}
	})
	if !found {
		t.Fatal("build info must retain the language profile")
	}
}

func TestTypefnRawAPIOptions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		input   string
		valid   bool
		enabled bool
	}{
		{`{"typefn":{"version":1,"limits":{"callDepth":16}}}`, true, true},
		{`{"typefn":{"version":2}}`, false, false},
		{`{"typefn":{"version":1,"limits":{"callDepth":0}}}`, false, false},
		{`{"typefn":{"version":1,"unknown":true}}`, false, false},
		{`{"typefn":false}`, false, false},
		{`{"typefn":null}`, true, false},
	} {
		t.Run(test.input, func(t *testing.T) {
			var raw RawCompilerOptions
			if err := json.Unmarshal([]byte(test.input), &raw); err != nil {
				t.Fatal(err)
			}
			options, errors := raw.Finalize("/")
			if (len(errors) == 0) != test.valid || (options.Typefn != nil) != test.enabled {
				t.Fatalf("options=%#v errors=%v", options.Typefn, errors)
			}
			if test.enabled && options.Typefn.Limits.CallDepth != 16 {
				t.Fatal("nested API limit was lost")
			}
		})
	}
}
