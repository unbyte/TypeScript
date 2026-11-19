package checker

import (
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/jsnum"
	"github.com/microsoft/TypeScript/tsc/internal/stringutil"
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
	record := c.newAnonymousType(nil, members, nil, nil, nil)
	root := c.typeFunctionExecution
	if root.records == nil {
		root.records = make(map[*Type]bool)
	}
	root.records[record] = true
	return record
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
	source := args[0]
	if op == "fromCodePoint" {
		n, ok := typeFunctionOffset(source)
		if !ok || n > 0x10ffff {
			return nil
		}
		return c.getStringLiteralType(stringutil.EncodeJSStringRune(rune(n)))
	}
	s, ok := typeFunctionString(source)
	if !ok {
		return nil
	}
	s = stringutil.CombineSurrogatePairs(s)
	pos := typeFunctionTextPosition{text: s}
	if op == "read" || op == "find" || op == "match" {
		index, ok := typeFunctionOffset(args[1])
		if !ok || !pos.seek(index) {
			return nil
		}
	}
	asString := c.getStringLiteralType
	matchedText := func(begin int) *Type {
		return c.createTupleType([]*Type{asString(s[begin:pos.bytes]), c.typeFunctionInt(pos.units)})
	}
	switch op {
	case "length":
		for pos.bytes < len(s) {
			pos.advance()
		}
		return c.typeFunctionInt(pos.units)
	case "read":
		if pos.bytes == len(s) {
			return c.undefinedType
		}
		before := pos.bytes
		pos.advance()
		return matchedText(before)
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
		needle, ok := typeFunctionString(args[2])
		if !ok {
			return nil
		}
		needle = stringutil.CombineSurrogatePairs(needle)
		i := strings.Index(s[pos.bytes:], needle)
		if i < 0 {
			return c.undefinedType
		}
		at := pos.bytes + i
		for pos.bytes < at {
			pos.advance()
		}
		begin := pos.units
		for pos.bytes < at+len(needle) {
			pos.advance()
		}
		return c.typeFunctionRecord(typeFunctionField{"source", source}, typeFunctionField{"start", c.typeFunctionInt(begin)}, typeFunctionField{"end", c.typeFunctionInt(pos.units)})
	case "match":
		pattern := args[2]
		begin := pos.bytes
		if exact, ok := typeFunctionString(pattern); ok {
			exact = stringutil.CombineSurrogatePairs(exact)
			if !strings.HasPrefix(s[begin:], exact) {
				return c.undefinedType
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
				if !c.typeFunctionCharge(node, len(ranges)) {
					return c.anyType
				}
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
				return c.undefinedType
			}
		}
		return matchedText(begin)
	}
	return nil
}
