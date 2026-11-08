package checker

import "github.com/microsoft/TypeScript/tsc/internal/ast"

// Definition checking follows lowered control flow without evaluating Types or
// executing loops. Both branch outcomes and zero-iteration loop exits are possible.
// Return and explicit failure terminate a path; a wildcard/capture cannot fail.
func (p *typefnProgram) canFallThrough() bool {
	seen := make([]bool, len(p.code)+1)
	work := []int{0}
	for len(work) != 0 {
		pc := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[pc] {
			continue
		}
		seen[pc] = true
		if pc == len(p.code) {
			return true
		}
		i := &p.code[pc]
		switch i.op {
		case tfReturn:
			continue
		case tfDiscard:
			if i.expr.intrinsic == "fail" || i.expr.intrinsic == "failAt" {
				continue
			}
		case tfJump:
			work = append(work, i.target)
			continue
		case tfBranch, tfIterNext:
			work = append(work, i.target)
		case tfMatch:
			if kind := i.pattern.node.Kind; kind != ast.KindTypeFunctionWildcardPattern && kind != ast.KindTypeFunctionCapturePattern {
				work = append(work, i.target)
			}
		}
		work = append(work, pc+1)
	}
	return false
}
