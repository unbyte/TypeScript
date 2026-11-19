package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
	"math"
	"strings"
)

type typeFunctionField struct {
	name  string
	value *Type
}

func (c *Checker) typeFunctionRecord(fields ...typeFunctionField) *Type {
	members := ast.NewSymbolTable()
	for _, field := range fields {
		members.Set(field.name, c.newProperty(field.name, field.value))
	}
	return c.newAnonymousType(nil, members, nil, nil, nil)
}
func (c *Checker) typeFunctionInt(n int) *Type { return c.getNumberLiteralType(jsnum.Number(n)) }
func typeFunctionOffset(t *Type) (int, bool) {
	n, ok := typeFunctionNumber(t)
	return int(n), ok && n >= 0 && n < math.MaxInt && math.Trunc(n) == n
}

type typeFunctionTextPosition struct {
	text         string
	bytes, units int
}

func (p *typeFunctionTextPosition) advance() rune {
	ch, size := stringutil.DecodeJSStringRune(p.text[p.bytes:])
	p.bytes += size
	p.units++
	if ch > 0xffff {
		p.units++
	}
	return ch
}
func (p *typeFunctionTextPosition) seek(offset int) bool {
	for p.units < offset && p.bytes < len(p.text) {
		p.advance()
	}
	return p.units == offset
}
func (c *Checker) typeFunctionText(node *ast.Node, op string, args []*Type) *Type {
	first := args[0]
	if op == "fromCodePoint" {
		n, ok := typeFunctionOffset(first)
		if !ok || n > 0x10ffff {
			return nil
		}
		return c.getStringLiteralType(stringutil.EncodeJSStringRune(rune(n)))
	}
	cursor := op == "read" || op == "find" || op == "match"
	source := first
	start := 0
	if cursor {
		if first.flags&TypeFlagsObject == 0 {
			return nil
		}
		source = c.getTypeOfPropertyOfType(first, "source")
		var ok bool
		start, ok = typeFunctionOffset(c.getTypeOfPropertyOfType(first, "offset"))
		if !ok {
			return nil
		}
	}
	s, ok := typeFunctionString(source)
	if !ok {
		return nil
	}
	s = stringutil.CombineSurrogatePairs(s)
	pos := typeFunctionTextPosition{text: s}
	if !pos.seek(start) {
		return nil
	}
	asString := c.getStringLiteralType
	tuple := func(parts ...*Type) *Type { return c.createTupleType(parts) }
	next := func() *Type {
		return c.typeFunctionRecord(typeFunctionField{"source", source}, typeFunctionField{"offset", c.typeFunctionInt(pos.units)})
	}
	switch op {
	case "length":
		for pos.bytes < len(s) {
			pos.advance()
		}
		return c.typeFunctionInt(pos.units)
	case "read":
		if pos.bytes == len(s) {
			return tuple(asString("eof"))
		}
		before := pos.bytes
		pos.advance()
		return tuple(asString("hit"), asString(s[before:pos.bytes]), next())
	case "codePoint":
		if len(s) == 0 {
			return nil
		}
		ch := pos.advance()
		if pos.bytes != len(s) {
			return nil
		}
		return c.typeFunctionInt(int(ch))
	case "slice":
		from, ok := typeFunctionOffset(args[1])
		to, ok2 := typeFunctionOffset(args[2])
		if !ok || !ok2 || to < from || !pos.seek(from) {
			return nil
		}
		i := pos.bytes
		if !pos.seek(to) {
			return nil
		}
		return asString(s[i:pos.bytes])
	case "split":
		sep, ok := typeFunctionString(args[1])
		if !ok || sep == "" {
			return nil
		}
		var result []*Type
		for part := range strings.SplitSeq(s, stringutil.CombineSurrogatePairs(sep)) {
			if !c.typeFunctionStep(node) {
				return c.anyType
			}
			result = append(result, asString(part))
		}
		return c.createTupleType(result)
	case "find":
		needle, ok := typeFunctionString(args[1])
		if !ok {
			return nil
		}
		needle = stringutil.CombineSurrogatePairs(needle)
		i := strings.Index(s[pos.bytes:], needle)
		if i < 0 {
			return tuple(asString("miss"))
		}
		at := pos.bytes + i
		for pos.bytes < at {
			pos.advance()
		}
		begin := pos.units
		for pos.bytes < at+len(needle) {
			pos.advance()
		}
		return tuple(asString("hit"), c.typeFunctionRecord(typeFunctionField{"source", source}, typeFunctionField{"start", c.typeFunctionInt(begin)}, typeFunctionField{"end", c.typeFunctionInt(pos.units)}))
	case "match":
		pattern := args[1]
		begin := pos.bytes
		if exact, ok := typeFunctionString(pattern); ok {
			exact = stringutil.CombineSurrogatePairs(exact)
			if !strings.HasPrefix(s[begin:], exact) {
				return tuple(asString("miss"))
			}
			for pos.bytes < begin+len(exact) {
				pos.advance()
			}
		} else {
			if pattern.flags&TypeFlagsObject == 0 {
				return nil
			}
			run := c.getTypeOfPropertyOfType(pattern, "run")
			if run == nil || !run.IsTupleType() || len(c.getTypeArguments(run)) == 0 {
				return nil
			}
			var ranges [][2]rune
			for _, pair := range c.getTypeArguments(run) {
				if !pair.IsTupleType() || len(c.getTypeArguments(pair)) != 2 {
					return nil
				}
				var interval [2]rune
				for i, t := range c.getTypeArguments(pair) {
					text, ok := typeFunctionString(t)
					if !ok || text == "" {
						return nil
					}
					text = stringutil.CombineSurrogatePairs(text)
					ch, size := stringutil.DecodeJSStringRune(text)
					if size != len(text) {
						return nil
					}
					interval[i] = ch
				}
				if interval[0] > interval[1] {
					return nil
				}
				ranges = append(ranges, interval)
			}
			minRun, maxRun := 1, math.MaxInt
			for key, target := range map[string]*int{"min": &minRun, "max": &maxRun} {
				if v := c.getTypeOfPropertyOfType(pattern, key); v != nil {
					value, ok := typeFunctionOffset(v)
					if !ok {
						return nil
					}
					*target = value
				}
			}
			if minRun > maxRun {
				return nil
			}
			count := 0
			for count < maxRun && pos.bytes < len(s) {
				after := pos
				ch := after.advance()
				match := false
				for _, r := range ranges {
					if r[0] <= ch && ch <= r[1] {
						match = true
						break
					}
				}
				if !match {
					break
				}
				pos = after
				count++
			}
			if count < minRun {
				return tuple(asString("miss"))
			}
		}
		return tuple(asString("hit"), asString(s[begin:pos.bytes]), next())
	}
	return nil
}
