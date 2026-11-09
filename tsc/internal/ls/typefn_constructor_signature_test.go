package ls

import "testing"

func TestTypefnConstructorBindingSignatureHelp(t *testing.T) {
	for _, test := range []struct {
		name, source, label string
		arity, active       int
	}{
		{"parameter", `typefn Apply<F<_>,T>{return F</*at*/T>;}`, "type F: typefn<_>", 1, 0},
		{"parameter-while-typing", `typefn Apply<F<_,_>,T>{return F<T,/*at*/`, "type F: typefn<_, _>", 2, 1},
		{"native-island", `typefn Apply<F<_,_>,T>{return type(F<[T,T],/*at*/T>);}`, "type F: typefn<_, _>", 2, 1},
		{"capture", `typefn Apply<F<_>,T>{typefn Inner<U>{return F</*at*/U>;} return Inner<T>;}`, "type F: typefn<_>", 1, 0},
		{"shadow", `typefn Apply<F<_>,T>{typefn Inner<F<_,_>,U>{return F<U,/*at*/U>;}return T;}`, "type F: typefn<_, _>", 2, 1},
		{"local", `typefn Use<T>{type F=typefn<A,B>{return [A,B];};return F<[T,T],/*at*/T>;}`, "type F: typefn<_, _>", 2, 1},
		{"reassigned", `typefn Use<T>{type F=typefn<A>{return [A];};F=typefn<B>{return B;};return F</*at*/T>;}`, "type F: typefn<_>", 1, 0},
		{"zero-arity", `typefn Use<>{type F=typefn<>{return 1;};return F</*at*/>;}`, "type F: typefn<>", 0, 0},
		{"does-not-execute", `typefn Fail<T>:type{fail<"must not execute for signature help">;}typefn Use<T>{type F=Fail;return F</*at*/T>;}`, "type F: typefn<_>", 1, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			l, ctx, pos := typefnService(t, test.source)
			for range 2 {
				response, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*at*/"), nil)
				if err != nil || response.SignatureHelp == nil {
					t.Fatalf("missing signature help: %#v %v", response, err)
				}
				help := response.SignatureHelp
				if len(help.Signatures) != 1 || help.Signatures[0].Label != test.label {
					t.Fatalf("signatures: %#v", help.Signatures)
				}
				parameters := help.Signatures[0].Parameters
				if parameters == nil || len(*parameters) != test.arity {
					t.Fatalf("parameters: %#v, want %d", parameters, test.arity)
				}
				for _, parameter := range *parameters {
					if parameter.Label.String == nil || *parameter.Label.String != "_" {
						t.Fatalf("parameter: %#v", parameter)
					}
				}
				if help.ActiveParameter == nil || help.ActiveParameter.Uinteger == nil || int(*help.ActiveParameter.Uinteger) != test.active {
					t.Fatalf("active parameter: %#v", help.ActiveParameter)
				}
				if len(l.program.GetSourceFile("/main.ts").Diagnostics()) == 0 {
					if diagnostics := l.program.GetSemanticDiagnostics(ctx, nil); len(diagnostics) != 0 {
						t.Fatalf("queries introduced errors: %v", diagnostics)
					}
				}
			}
		})
	}
}

func TestTypefnTypeBindingsHaveNoConstructorSignature(t *testing.T) {
	for _, source := range []string{
		`typefn Use<T>{return T</*at*/`,
		`typefn Use<T>{type F=T;return F</*at*/T>;}`,
	} {
		t.Run(source, func(t *testing.T) {
			l, ctx, pos := typefnService(t, source)
			response, err := l.ProvideSignatureHelp(ctx, "file:///main.ts", pos("/*at*/"), nil)
			if err != nil || response.SignatureHelp != nil {
				t.Fatalf("non-constructor binding acquired signature help: %#v %v", response, err)
			}
		})
	}
}
