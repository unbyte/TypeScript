package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"unsafe"
)

type typefnTemplatePatternPart struct {
	text    []uint16
	capture *typefnPattern
}

// A cursor over a normalized native template. Holes are opaque Types, never
// enumerated strings; only a final capture can consume an opaque suffix.
type typefnPatternCursor struct {
	texts         [][]uint16
	holes         []*Type
	index, offset int
}

func (v *typefnVM) templatePattern(p *typefnPattern, subject *Type, slots []*Type) bool {
	node := p.node
	if subject.flags != TypeFlagsStringLiteral && subject.flags != TypeFlagsString && subject.flags != TypeFlagsTemplateLiteral {
		if typefnUnresolved(subject) {
			v.blocked = true
			return false
		}
		if subject.flags&TypeFlagsUnion != 0 || subject.flags&TypeFlagsIntersection != 0 && v.c.isTypeAssignableTo(subject, v.c.stringType) {
			v.error(node, "template patterns require one unbranded string subject")
		}
		return false
	}
	// Evaluate all pins before attempting chunks; a mismatching prefix cannot
	// conceal an invalid pin later in this arm. Captures have not been published.
	expression := node.AsTypeFunctionPattern().Expression
	maxParts := 1 + 2*len(p.children)
	subjectTexts := 2
	if subject.flags == TypeFlagsTemplateLiteral {
		subjectTexts = len(subject.AsTemplateLiteralType().texts)
	}
	bookkeeping := maxParts*2*int(unsafe.Sizeof(typefnTemplatePatternPart{})) + subjectTexts*int(unsafe.Sizeof([]uint16{})+unsafe.Sizeof(""))
	if bookkeeping > v.limits.LiveBytes-v.liveBytes {
		v.error(node, "live VM allocation limit exceeded")
		return false
	}
	v.liveBytes += bookkeeping
	defer func() { v.liveBytes -= bookkeeping }()
	raw := make([]typefnTemplatePatternPart, 0, maxParts)
	appendText := func(text string) bool {
		units, ok := v.stringUnitsOf(node, v.c.getStringLiteralType(text))
		if ok {
			raw = append(raw, typefnTemplatePatternPart{text: units})
		}
		return ok
	}
	if expression.Kind == ast.KindTypeFunctionTemplate {
		template := expression.AsTypeFunctionTemplate()
		if !appendText(template.Head.Text()) {
			return false
		}
		for i, child := range p.children {
			if child.node.Kind == ast.KindTypeFunctionPinPattern {
				value := v.expression(child.expr, slots)
				if v.failed || v.blocked {
					return false
				}
				if typefnUnresolved(value) {
					v.blocked = true
					return false
				}
				units, ok := v.stringUnitsOf(child.node, value)
				if !ok {
					return false
				}
				if len(units) == 0 {
					v.error(child.node, "template pin requires a nonempty string literal")
					return false
				}
				raw = append(raw, typefnTemplatePatternPart{text: units})
			} else {
				raw = append(raw, typefnTemplatePatternPart{capture: child})
			}
			if !appendText(template.Spans.Nodes[i].AsTypeFunctionTemplateSpan().Literal.Text()) {
				return false
			}
		}
	} else {
		literal := expression.AsTypeFunctionExpression().NativeType.AsLiteralTypeNode().Literal
		if !appendText(literal.Text()) {
			return false
		}
	}
	// Join neighboring literal/pin chunks once, so a capture searches for the
	// entire delimiter. In particular a pin is not an extra capture boundary.
	units := 0
	for _, part := range raw {
		if len(part.text) > v.limits.StringUnits-v.stringUnits-units {
			v.error(node, "string work limit exceeded")
			return false
		}
		units += len(part.text)
	}
	if units > (v.limits.LiveBytes-v.liveBytes)/2 {
		v.error(node, "live VM allocation limit exceeded")
		return false
	}
	v.liveBytes += units * 2
	defer func() { v.liveBytes -= units * 2 }()
	if !v.stringWork(node, units) {
		return false
	}
	parts := make([]typefnTemplatePatternPart, 0, len(raw))
	for i := 0; i < len(raw); {
		if raw[i].capture != nil {
			parts = append(parts, raw[i])
			i++
			continue
		}
		end, size := i, 0
		for end < len(raw) && raw[end].capture == nil {
			size += len(raw[end].text)
			end++
		}
		if size != 0 {
			joined := make([]uint16, 0, size)
			for ; i < end; i++ {
				joined = append(joined, raw[i].text...)
			}
			parts = append(parts, typefnTemplatePatternPart{text: joined})
		} else {
			i = end
		}
	}
	cursor := typefnPatternCursor{}
	texts := []string{"", ""}
	switch subject.flags {
	case TypeFlagsStringLiteral:
		texts = []string{subject.AsLiteralType().value.(string)}
	case TypeFlagsString:
		cursor.holes = []*Type{subject}
	case TypeFlagsTemplateLiteral:
		template := subject.AsTemplateLiteralType()
		texts, cursor.holes = template.texts, template.types
	}
	cursor.texts = make([][]uint16, 0, len(texts))
	for _, text := range texts {
		units, ok := v.stringUnitsOf(node, v.c.getStringLiteralType(text))
		if !ok {
			return false
		}
		cursor.texts = append(cursor.texts, units)
	}
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if part.capture == nil {
			if !v.patternPrefix(node, &cursor, part.text) {
				return false
			}
			continue
		}
		if i == len(parts)-1 {
			if part.capture.node.Kind == ast.KindTypeFunctionWildcardPattern {
				return true
			}
			value := v.patternSuffix(node, &cursor)
			if v.failed {
				return false
			}
			slots[part.capture.slot] = value
			return true
		}
		available := cursor.texts[cursor.index][cursor.offset:]
		count := 1 // Adjacent captures consume one UTF-16 unit except the final one.
		if parts[i+1].capture == nil {
			needle := parts[i+1].text
			prefix, release, ok := v.textSearcher(node, needle)
			if !ok {
				return false
			}
			count = v.textIndex(node, available, needle, prefix, 0)
			release()
			if v.failed {
				return false
			}
			if count < 0 {
				v.patternOpaque(&cursor)
				return false
			}
		} else if len(available) == 0 {
			v.patternOpaque(&cursor)
			return false
		}
		if part.capture.node.Kind != ast.KindTypeFunctionWildcardPattern {
			value := v.textResult(node, available[:count])
			if v.failed {
				return false
			}
			slots[part.capture.slot] = value
		}
		cursor.offset += count
	}
	if cursor.index == len(cursor.holes) && cursor.offset == len(cursor.texts[cursor.index]) {
		return true
	}
	// A symbolic remaining hole may later become empty. A known nonempty literal
	// suffix already disproves the anchored match without observing the hole.
	if cursor.offset == len(cursor.texts[cursor.index]) {
		v.patternOpaque(&cursor)
	}
	return false
}

func (v *typefnVM) patternOpaque(cursor *typefnPatternCursor) {
	if cursor.index < len(cursor.holes) && typefnUnresolved(cursor.holes[cursor.index]) {
		v.blocked = true
	}
}

func (v *typefnVM) patternPrefix(node *ast.Node, cursor *typefnPatternCursor, prefix []uint16) bool {
	available := cursor.texts[cursor.index][cursor.offset:]
	for index, unit := range prefix {
		if index >= len(available) {
			v.patternOpaque(cursor)
			return false
		}
		if !v.stringWork(node, 1) || available[index] != unit {
			return false
		}
	}
	cursor.offset += len(prefix)
	return true
}

func (v *typefnVM) patternSuffix(node *ast.Node, cursor *typefnPatternCursor) *Type {
	texts := make([]string, len(cursor.texts)-cursor.index)
	for i := cursor.index; i < len(cursor.texts); i++ {
		units := cursor.texts[i]
		if i == cursor.index {
			units = units[cursor.offset:]
		}
		value := v.textResult(node, units)
		if v.failed {
			return v.c.errorType
		}
		texts[i-cursor.index] = value.AsLiteralType().value.(string)
	}
	if !v.allocate(node, len(texts)) {
		return v.c.errorType
	}
	return v.c.getTemplateLiteralType(texts, cursor.holes[cursor.index:])
}
