package typefntest

import (
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

// The oracle evaluates generated trees with math/big. It neither parses the
// rendered input nor reproduces the DSL's precedence-climbing implementation.
type calculatorExpr struct {
	op          byte
	literal     string
	left, right *calculatorExpr
}

type calculatorFailure struct {
	expr    *calculatorExpr
	message string
}

func (e *calculatorExpr) evaluate() (*big.Int, *calculatorFailure) {
	if e.op == 0 {
		n, ok := new(big.Int).SetString(e.literal, 10)
		if !ok {
			panic("invalid generated integer")
		}
		return n, nil
	}
	left, failure := e.left.evaluate()
	if failure != nil {
		return nil, failure
	}
	if e.right == nil {
		if e.op == '-' {
			left.Neg(left)
		}
		return left, nil
	}
	right, failure := e.right.evaluate()
	if failure != nil {
		return nil, failure
	}
	result := new(big.Int)
	switch e.op {
	case '+':
		result.Add(left, right)
	case '-':
		result.Sub(left, right)
	case '*':
		result.Mul(left, right)
	case '/':
		if right.Sign() == 0 {
			return nil, &calculatorFailure{e, "Division by zero"}
		}
		remainder := new(big.Int)
		result.QuoRem(left, right, remainder)
		if remainder.Sign() != 0 {
			return nil, &calculatorFailure{e, "Non-integral division"}
		}
	default:
		panic("invalid generated operator")
	}
	return result, nil
}

func (e *calculatorExpr) precedence() int {
	if e.op == 0 {
		return 4
	}
	if e.right == nil {
		return 3
	}
	if e.op == '*' || e.op == '/' {
		return 2
	}
	return 1
}

// Minimal parentheses preserve the tree, including right-associated division
// and subtraction. A second rendering adds redundant groups and whitespace.
// Offsets are measured while rendering, independently of the evaluator. All
// generated characters are ASCII, so byte and UTF-16 offsets coincide.
func (e *calculatorExpr) render(grouped bool) (string, map[*calculatorExpr]int) {
	var out strings.Builder
	offsets := map[*calculatorExpr]int{}
	space := ""
	if grouped {
		space = " \t\r\n"
	}
	var write func(*calculatorExpr, int)
	write = func(node *calculatorExpr, minimum int) {
		paren := node.precedence() < minimum || grouped && node.op != 0
		if paren {
			out.WriteString("(" + space)
		}
		if node.op == 0 {
			out.WriteString(node.literal)
		} else if node.right == nil {
			offsets[node] = out.Len()
			out.WriteByte(node.op)
			out.WriteString(space)
			write(node.left, node.precedence())
		} else {
			write(node.left, node.precedence())
			out.WriteString(space)
			offsets[node] = out.Len()
			out.WriteByte(node.op)
			out.WriteString(space)
			write(node.right, node.precedence()+1)
		}
		if paren {
			out.WriteString(space + ")")
		}
	}
	out.WriteString(space)
	write(e, 0)
	out.WriteString(space)
	return out.String(), offsets
}

func generatedCalculatorTrees() []*calculatorExpr {
	rng := rand.New(rand.NewPCG(0x74797065666e, 20))
	literals := []string{"0", "1", "2", "3", "7", "10", "99", "00012", "9007199254740993", "18446744073709551616", "123456789012345678901234567890"}
	var generate func(int) *calculatorExpr
	generate = func(depth int) *calculatorExpr {
		if depth == 0 || rng.IntN(5) == 0 {
			return &calculatorExpr{literal: literals[rng.IntN(len(literals))]}
		}
		if rng.IntN(4) == 0 {
			return &calculatorExpr{op: "+-"[rng.IntN(2)], left: generate(depth - 1)}
		}
		return &calculatorExpr{op: "+-*/"[rng.IntN(4)], left: generate(depth - 1), right: generate(depth - 1)}
	}
	var trees []*calculatorExpr
	for i := range 96 {
		trees = append(trees, generate(1+i%4))
	}
	// Ensure exact divisions, signed divisors, zero divisors and remainders are
	// represented regardless of changes to the random corpus.
	for _, sign := range []byte{'+', '-'} {
		for _, divisorSign := range []byte{'+', '-'} {
			for _, numerator := range []string{"0", "7", "14", "18446744073709551616"} {
				for _, divisor := range []string{"0", "2", "7"} {
					trees = append(trees, &calculatorExpr{op: '/',
						left:  &calculatorExpr{op: sign, left: &calculatorExpr{literal: numerator}},
						right: &calculatorExpr{op: divisorSign, left: &calculatorExpr{literal: divisor}}})
				}
			}
		}
	}
	return trees
}

func TestTypefnGeneratedCalculator(t *testing.T) {
	const header = `import type { Calc, CalcBig } from "./calculator";
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
`
	var successes, failures strings.Builder
	successes.WriteString(header)
	failures.WriteString(header)
	var diagnostics []wantDiagnostic
	counts := map[string]int{}
	for i, tree := range generatedCalculatorTrees() {
		value, failure := tree.evaluate()
		for variant, grouped := range []bool{false, true} {
			input, offsets := tree.render(grouped)
			// A comment makes duplicate generated expressions distinct use sites.
			argument := fmt.Sprintf("%s /* case %d/%d */", strconv.Quote(input), i, variant)
			application := "CalcBig<" + argument + ">"
			if failure != nil {
				fmt.Fprintf(&failures, "type Failure%d_%d = %s;\n", i, variant, application)
				diagnostics = append(diagnostics, wantDiagnostic{Phase: "semantic", Code: -1, File: "assertions.ts", At: application,
					Message: fmt.Sprintf("%s (UTF-16 offset %d)", failure.message, offsets[failure.expr])})
				counts[failure.message]++
			} else {
				fmt.Fprintf(&successes, "type Value%d_%d = Assert<Equal<%s, %sn>>;\n", i, variant, application, value.String())
				counts["value"]++
				if value.IsInt64() && value.Int64() >= -9007199254740991 && value.Int64() <= 9007199254740991 {
					fmt.Fprintf(&successes, "type Number%d_%d = Assert<Equal<Calc<%s>, %s>>;\n", i, variant, argument, value.String())
					counts["number"]++
				}
			}
		}
	}
	for _, category := range []string{"value", "number", "Division by zero", "Non-integral division"} {
		if counts[category] == 0 {
			t.Fatalf("generated corpus omitted %s", category)
		}
	}
	t.Logf("generated calculator assertions: %v", counts)
	for _, test := range []struct {
		name, source string
		diagnostics  []wantDiagnostic
	}{
		{"values", successes.String(), nil},
		{"failures", failures.String(), diagnostics},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := fixture{
				Files:      map[string]string{"calculator.ts": "../../../../docs/typefn/examples/calculator.typefn", "assertions.ts": ""},
				sourceText: map[string]string{"assertions.ts": test.source},
				Options:    map[string]any{"typefn": map[string]any{"version": 1}}, Diagnostics: test.diagnostics,
			}
			t.Run("source", func(t *testing.T) { runFixture(t, f) })
			for _, skip := range []bool{false, true} {
				t.Run(fmt.Sprintf("declaration/skipLibCheck=%v", skip), func(t *testing.T) {
					consumer := f
					consumer.Files = map[string]string{"assertions.ts": ""}
					consumer.sourceText = map[string]string{"assertions.ts": strings.Replace(test.source, `"./calculator"`, `"./out/calculator"`, 1)}
					consumer.Options = map[string]any{"typefn": map[string]any{"version": 1}, "skipLibCheck": skip}
					runFixture(t, fixture{
						Files:    map[string]string{"calculator.ts": f.Files["calculator.ts"]},
						Options:  map[string]any{"typefn": map[string]any{"version": 1}, "declaration": true, "emitDeclarationOnly": true},
						Consumer: &consumer,
					})
				})
			}
		})
	}
}
