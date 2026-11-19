// @strict: true
// @target: esnext

typefn Failed<> { fail<"inner failure">; }
typefn ExplicitAny<> { return any; }
typefn After<T> { type A = T; return 1; }
typefn Outer<> { return { version: 1; bad: Failed<> }; }
typefn First<T> { return 42; }
type Open<T extends unknown[]> = First<[42, ...T]>;
typefn Identity<T> { return T; }
type Unused = Failed<>;
function same<T>(x: Identity<T>): Identity<T> { return x; }
// @ts-expect-error an opaque generic computation cannot promise string
function unsupported<T>(x: Identity<T>): string { return x; }

declare const good: After<ExplicitAny<>>;
const literal: 1 = good;
declare const bad: Outer<>;
bad.anyPropertyAfterRecovery;

typefn Snapshots<> {
    type X = "before";
    type Saved = { read: () => X };
    X = "after";
    return [Saved, X];
}
declare const snapshot: Snapshots<>;
const before: "before" = snapshot[0].read();
const after: "after" = snapshot[1];
