// @strict: true
// @target: esnext

typefn Walk<N> {
    if (N === 0) { return 0; }
    return Walk<number.subtract<N, 1>>;
}
typefn Spin<> { while (true) {} }
typefn Constant<T> { return 7; }

declare const shallow: Walk<8>;
const zero: 0 = shallow;
declare const deep: Walk<150>;
deep.recovered;
declare const spin: Spin<>;
spin.recovered;

// Rebuilding an object charges the fields copied, even with a short loop.
type Digit = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;
type Wide = { [K in `k${Digit}${Digit}${Digit}`]: 1 };
typefn ReplaceMany<T> {
    type Out = T;
    type I = 0;
    while (number.lessThan<I, 128>) {
        Out = object.setField<Out, "extra", { type: 2 }>;
        I = number.add<I, 1>;
    }
    return Out;
}
declare const wide: ReplaceMany<Wide>;
wide.recovered;

// A failed root cannot spend the next application's budget.
declare const next: Constant<"ok">;
const seven: 7 = next;
