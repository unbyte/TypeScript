package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func TestTypefnCallDefaultPublication(t *testing.T) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.ts", PathKey: "/test.ts"}, "select();", core.ScriptKindTS)
	node := file.Statements.Nodes[0].AsExpressionStatement().Expression
	c := &Checker{}
	signature := &Signature{}
	defaults := &InferenceContext{inferences: []*InferenceInfo{{fromDefault: true}}}
	inferred := &InferenceContext{inferences: []*InferenceInfo{{}}}
	publish := func(context *InferenceContext) *typefnCallDefaults {
		frame := c.beginTypefnCallResolution(node)
		c.recordTypefnCallDefaults(node, signature, context)
		c.publishTypefnCallResolution(frame, signature, false)
		c.endTypefnCallResolution(frame)
		return c.typefnCallDefaultOrigin(node)
	}
	original := publish(defaults)
	if original == nil || !original.used[0] {
		t.Fatal("missing committed default")
	}

	// Nested resolution can commit the same signature with different provenance.
	// The outer result must not overwrite it, even after stack storage grows.
	outer := c.beginTypefnCallResolution(node)
	for range 100 {
		c.beginTypefnCallResolution(node)
	}
	nested := publish(defaults)
	if nested == original {
		t.Fatal("nested resolution did not publish fresh provenance")
	}
	for i := 100; i > 0; i-- {
		c.endTypefnCallResolution(i)
	}
	c.recordTypefnCallDefaults(node, signature, inferred)
	c.publishTypefnCallResolution(outer, signature, true)
	c.endTypefnCallResolution(outer)
	if c.typefnCallDefaultOrigin(node) != nested {
		t.Fatal("outer resolution overwrote the nested commit")
	}

	// A transient control-flow/speculative result must not change published facts.
	frame := c.beginTypefnCallResolution(node)
	c.recordTypefnCallDefaults(node, signature, inferred)
	c.endTypefnCallResolution(frame)
	if c.typefnCallDefaultOrigin(node) != nested {
		t.Fatal("unpublished resolution changed provenance")
	}

	// An early error-signature cache write is not a nested provenance publication.
	frame = c.beginTypefnCallResolution(node)
	c.publishTypefnCallResolution(frame, signature, true)
	c.endTypefnCallResolution(frame)
	if c.typefnCallDefaultOrigin(node) != nil {
		t.Fatal("error signature retained stale default facts")
	}

	original = publish(defaults)
	if publish(inferred) != nil {
		t.Fatal("inference retained a prior default with the same signature")
	}
	c.restoreTypefnCallDefaultOrigin(node, original)
	if c.typefnCallDefaultOrigin(node) != original {
		t.Fatal("query restoration lost committed facts")
	}
	c.restoreTypefnCallDefaultOrigin(node, nil)
	if c.typefnCallDefaultOrigin(node) != nil || len(c.typefn.callResolutions) != 0 {
		t.Fatal("query cleanup retained temporary state")
	}
}
