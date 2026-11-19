//// [tests/cases/compiler/typeFunctionOperators.ts] ////

//// [typeFunctionOperators.ts]
type Identity<T> = T;
typefn Id<T> { return T; }
typefn Failed<> { fail<"unexpected evaluation">; }
export typefn Relations<T, U> {
    return [T is U, T === U, T !== U, T is string | number && !(T === never)];
}
export typefn Arithmetic<A, B> {
    return [A + B * 3, (A + B) * 3, A - B - 1, -A, -(-A), A / B, A % B];
}
export typefn Nested<N> {
    type X = N;
    type Saved = { value: Id<X + 1>; enabled: Identity<true || Failed<>> };
    X = X * 2;
    return [Saved, Id<-X>, false && Failed<>, -1 | 1, true | false, true || false];
}
export typefn Predicate<T> { return (x: unknown) => x is T; }

declare const a: Relations<"x" | 1, string>;
const mixed: [false, false, true, true] = a;
declare const b: Relations<{ id: 1; name: "x" }, { id: number }>;
const structural: [true, false, true, false] = b;
declare const c: Relations<never, string>;
const bottom: [true, false, true, false] = c;
declare const d: Relations<any, string>;
const top: [true, false, true, true] = d;
declare const e: Relations<{ a: 1 } & { b: 2 }, { a: 1 } & { b: 2 }>;
const identity: [true, true, false, false] = e;
declare const numbers: Arithmetic<8, 2>;
const precedence: [14, 30, 5, -8, 8, 4, 0] = numbers;
typefn Big<A, B> { return [A + B, A - B, A * B, A / B, A % B, -A]; }
declare const big: Big<-7n, 3n>;
const integers: [-4n, -10n, -21n, -2n, -1n, 7n] = big;
declare const nested: Nested<4>;
const snapshot: [{ value: 5; enabled: true }, -8, false, -1 | 1, boolean, true] = nested;
declare const pred: Predicate<number>;
const guard: (x: unknown) => x is number = pred;

typefn Choose<T> { return T extends string ? (true || Failed<>) : (1 + 1); }
declare const chosen: Choose<"x">;
const yes: true = chosen;

type Box<T> = { value: T };
typefn Constructor<F<_>, T> { return [F<T> is Box<string>]; }
declare const constructed: Constructor<Box<_>, "x">;
const fits: [true] = constructed;
typefn NativeSignature<N> { return <T>(arg: T) => [T extends string ? true : false, N + 1]; }
declare const deferred: NativeSignature<3>;
const later: [true, 4] = deferred("hello");
const again: [false, 4] = deferred(0);

typefn Short<S> { return Id<true || Failed<>["x"]>; }
declare const short: Short<0>;
const skipped: true = short;


//// [typeFunctionOperators.js]
const mixed = a;
const structural = b;
const bottom = c;
const top = d;
const identity = e;
const precedence = numbers;
const integers = big;
const snapshot = nested;
const guard = pred;
const yes = chosen;
const fits = constructed;
const later = deferred("hello");
const again = deferred(0);
const skipped = short;
export {};


//// [typeFunctionOperators.d.ts]
type Identity<T> = T;
typefn Id<T> {
    return T;
}
typefn Failed<> {
    fail<"unexpected evaluation">;
}
export typefn Relations<T, U> {
    return [T is U, T === U, T !== U, T is string | number && !(T === never)];
}
export typefn Arithmetic<A, B> {
    return [A + B * 3, (A + B) * 3, A - B - 1, -A, -(-A), A / B, A % B];
}
export typefn Nested<N> {
    type X = N;
    type Saved = {
        value: Id<X + 1>;
        enabled: Identity<true || Failed<>>;
    };
    X = X * 2;
    return [Saved, Id<-X>, false && Failed<>, -1 | 1, true | false, true || false];
}
export typefn Predicate<T> {
    return (x: unknown) => x is T;
}
export {};
