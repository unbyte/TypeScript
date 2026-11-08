package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

func (v *typefnVM) templateExpression(e *typefnExpression, slots []*Type) *Type {
	node := e.node
	template := node.AsTypeFunctionTemplate()
	texts := make([]string, len(e.args)+1)
	values := make([]*Type, len(e.args))
	texts[0] = template.Head.Text()
	pending := false
	for i, arg := range e.args {
		values[i] = v.expression(arg, slots)
		if v.failed || v.blocked {
			return v.c.errorType
		}
		pending = pending || typefnUnresolved(values[i])
		if !v.c.isTypeAssignableTo(values[i], v.c.templateConstraintType) {
			if typefnUnresolved(values[i]) {
				pending = true
			} else {
				return v.error(arg.node, "template interpolation requires string, number, bigint, boolean, null or undefined")
			}
		}
		texts[i+1] = template.Spans.Nodes[i].AsTypeFunctionTemplateSpan().Literal.Text()
	}
	// Bound both the number and total known text size of native expansions before
	// normalization. Replayed checks run before guarded shape instantiation.
	count, units, ok := v.templateExpansion(node, texts, values)
	if !ok || !v.allocate(node, count+1) || !v.stringWork(node, units) {
		return v.c.errorType
	}
	result := v.c.getTemplateLiteralType(texts, values)
	if v.c.isErrorType(result) {
		v.failed = true
		return result
	}
	if pending {
		// Publish the native template shape, but retain eligibility checks so an
		// absorbed/projected result cannot erase an invalid later substitution.
		result = v.c.guardTypefn(result, []*Type{v.deferExpression(e, slots)})
	}
	return result
}

func (v *typefnVM) templateExpansion(node *ast.Node, texts []string, values []*Type) (int, int, bool) {
	count, units := 1, 0
	for _, text := range texts {
		decoded, ok := v.stringUnitsOf(node, v.c.getStringLiteralType(text))
		if !ok {
			return 0, 0, false
		}
		if len(decoded) > v.limits.StringUnits-units {
			v.error(node, "string work limit exceeded")
			return 0, 0, false
		}
		units += len(decoded)
	}
	for _, value := range values {
		n, size, ok := v.templateOperandExpansion(node, value)
		if !ok {
			return 0, 0, false
		}
		if n == 0 {
			return 0, 0, true
		}
		if count > (v.limits.NativeNodes-v.nodes)/n {
			v.error(node, "native construction limit exceeded")
			return 0, 0, false
		}
		budget := v.limits.StringUnits - v.stringUnits
		if units > budget/n || size > (budget-units*n)/count {
			v.error(node, "string work limit exceeded")
			return 0, 0, false
		}
		units = units*n + size*count
		count *= n
	}
	return count, units, true
}

func (v *typefnVM) templateOperandExpansion(node *ast.Node, value *Type) (int, int, bool) {
	if !v.step(node) {
		return 0, 0, false
	}
	if value.flags&TypeFlagsNever != 0 {
		return 0, 0, true
	}
	if value.flags&TypeFlagsUnion != 0 {
		count, units := 0, 0
		for _, part := range value.Types() {
			n, size, ok := v.templateOperandExpansion(node, part)
			if !ok {
				return 0, 0, false
			}
			if n > v.limits.NativeNodes-v.nodes-count {
				v.error(node, "native construction limit exceeded")
				return 0, 0, false
			}
			if size > v.limits.StringUnits-v.stringUnits-units {
				v.error(node, "string work limit exceeded")
				return 0, 0, false
			}
			count += n
			units += size
		}
		return count, units, true
	}
	if value.flags&TypeFlagsTemplateLiteral != 0 {
		t := value.AsTemplateLiteralType()
		return v.templateExpansion(node, t.texts, t.types)
	}
	if value.flags&(TypeFlagsLiteral|TypeFlagsNull|TypeFlagsUndefined) != 0 {
		text := v.c.getTemplateStringForType(value)
		units, ok := v.stringUnitsOf(node, v.c.getStringLiteralType(text))
		return 1, len(units), ok
	}
	return 1, 0, true // Native symbolic/pattern holes have no known concrete text.
}
