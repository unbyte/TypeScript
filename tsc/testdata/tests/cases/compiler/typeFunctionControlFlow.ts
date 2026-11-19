// @strict: true
// @target: esnext

typefn Sum<Items extends readonly number[]> {
    type Total = 0;
    for (type Item of Items) {
        if (Item === 0) { continue; }
        Total = number.add<Total, Item>;
    }
    return Total;
}

typefn Countdown<N> {
    type Remaining = N;
    while (number.lessThan<0, Remaining>) {
        Remaining = number.subtract<Remaining, 1>;
    }
    return Remaining;
}

typefn Classify<T> {
    match T {
        case [string, number] { return "pair"; }
        else { return "other"; }
    }
}
typefn Problem<> { fail<"deliberate">; }
typefn ShortCircuit<> {
    if (true || Problem<>) { return 1; }
    return 2;
}
declare const sum: Sum<[1, 0, 2]>;
declare const countdown: Countdown<4>;
declare const pair: Classify<["ok", 42]>;
declare const short: ShortCircuit<>;
const three: 3 = sum;
const zero: 0 = countdown;
const label: "pair" = pair;
const one: 1 = short;
// @ts-expect-error
const wrong: 6 = sum;
