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

typefn Selected<T> { return T extends "yes" ? 1 : Failed<>; }
declare const selected: Selected<"yes">;
const one: 1 = selected;

typefn Span<S> { fail<"unexpected token", { source: S; start: 2; end: 3 }>; }
declare const spelling: Span<"\u{1F600}\x2b">;
spelling.recovered;
type Computed = `😀${"+"}`;
declare const computed: Span<Computed>;
computed.recovered;

type FailedAlias = Failed<>;
declare const firstFailed: FailedAlias;
declare const secondFailed: FailedAlias;
firstFailed.a;
secondFailed.b;

typefn ConditionalField<T> {
    return { payload: T extends [infer I] ? { nested: Failed<> } : "ok"; stable: 1 };
}
declare const matchFails: ConditionalField<[1]>;
matchFails.afterFailure;
declare const matchSkips: ConditionalField<"outside">;
const skipped: "ok" = matchSkips.payload;
