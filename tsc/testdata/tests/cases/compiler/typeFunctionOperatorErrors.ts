// @strict: true
// @target: esnext

typefn Sum<A, B> { return A + B; }
typefn Negate<T> { return -T; }
typefn Divide<A, B> { return A / B; }
typefn Remainder<A, B> { return A % B; }
typefn Overflow<N> { return N * N; }
typefn And<A, B> { return A && B; }
declare const mixed: Sum<1, 2n>;
declare const text: Sum<"1", "2">;
declare const broad: Negate<number>;
declare const zero: Divide<3, 0>;
declare const bigZero: Remainder<3n, 0n>;
declare const infinite: Overflow<1e300>;
declare const truthy: And<true, 1>;
mixed.recovered; text.recovered; broad.recovered; zero.recovered; bigZero.recovered; infinite.recovered; truthy.recovered;
typefn Ungrouped<T> { return T is string === true; }
typefn Grouped<T> { return (T is string) === true; }
declare const ok: Grouped<"x">;
const yes: true = ok;

typefn Unexpected<S> { fail<"unexpected token", { source: S; start: 0; end: 1 }>; }
typefn CheckInput<> { return <S extends string>(input: S) => [false || Unexpected<S>]; }
declare const check: CheckInput<>;
const delayed = check("x");
delayed[0].recovered;
