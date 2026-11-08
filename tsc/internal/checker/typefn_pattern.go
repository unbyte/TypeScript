package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

// A false result is a definite non-match only when neither blocked nor failed
// was set. Registers for captures are arm-private and cannot leak to later arms.
func (v *typefnVM) pattern(p *typefnPattern, subject *Type, slots []*Type) bool {
	node := p.node
	if !v.step(node) {
		return false
	}
	subject = v.observe(subject)
	switch node.Kind {
	case ast.KindTypeFunctionTemplatePattern:
		return v.templatePattern(p, subject, slots)
	case ast.KindTypeFunctionWildcardPattern:
		return true
	case ast.KindTypeFunctionCapturePattern:
		slots[p.slot] = subject
		return true
	case ast.KindTypeFunctionLiteralPattern:
		value := v.expression(p.expr, slots)
		if v.failed || v.blocked {
			return false
		}
		if typefnUnresolved(subject) {
			v.blocked = true
			return false
		}
		return v.c.isTypeIdenticalTo(subject, value)
	case ast.KindTypeFunctionIsPattern, ast.KindTypeFunctionNativePattern:
		value := v.expression(p.expr, slots)
		if v.failed || v.blocked {
			return false
		}
		op := ast.KindEqualsEqualsToken
		if node.Kind == ast.KindTypeFunctionIsPattern {
			op = ast.KindIsKeyword
		}
		result := v.binary(node, op, subject, value)
		if v.failed {
			return false
		}
		truth, ok := v.boolean(node, result)
		return ok && truth
	case ast.KindTypeFunctionTuplePattern, ast.KindTypeFunctionObjectPattern:
		if typefnUnresolved(subject) || v.c.isGenericObjectType(subject) {
			v.blocked = true
			return false
		}
		if subject.flags&TypeFlagsUnion != 0 {
			v.error(node, "structural patterns require one subject, not a union")
			return false
		}
		if node.Kind == ast.KindTypeFunctionTuplePattern {
			if !isTupleType(subject) || subject.TargetTupleType().combinedFlags & ^ElementFlagsRequired != 0 {
				return false
			}
			elements := v.c.getTypeArguments(subject)
			count := len(p.children)
			rest := count > 0 && p.children[count-1].node.Kind == ast.KindTypeFunctionRestPattern
			if rest {
				count--
			}
			if len(elements) < count || !rest && len(elements) != count {
				return false
			}
			for index := 0; index < count; index++ {
				if !v.pattern(p.children[index], elements[index], slots) {
					return false
				}
			}
			if rest {
				tail := p.children[count].children[0]
				if tail.node.Kind == ast.KindTypeFunctionWildcardPattern {
					return true
				}
				if !v.allocate(node, len(elements)-count+1) {
					return false
				}
				target := subject.TargetTupleType()
				value := v.c.createTupleTypeEx(elements[count:], target.elementInfos[count:], target.readonly)
				return v.pattern(tail, value, slots)
			}
			return true
		}
		if subject.flags&(TypeFlagsObject|TypeFlagsIntersection) == 0 {
			return false
		}
		for _, field := range p.children {
			if !v.step(field.node) {
				return false
			}
			property := v.c.getPropertyOfTypeEx(subject, field.key, true, false)
			if property == nil || property.Flags()&ast.SymbolFlagsOptional != 0 || getDeclarationModifierFlagsFromSymbol(property)&ast.ModifierFlagsNonPublicAccessibilityModifier != 0 {
				return false
			}
			if !v.pattern(field.children[0], v.c.getNonMissingTypeOfSymbol(property), slots) {
				return false
			}
		}
		return true
	}
	return false
}
