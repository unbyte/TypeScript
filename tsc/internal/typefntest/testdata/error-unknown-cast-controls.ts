typefn Increment<N> { return N + 1; }
declare const value:unknown;
const invalid = value as Increment<number>;
function assignment<N>(value:unknown) { const result:Increment<N> = value; }
function check<N>(value:unknown) { return value satisfies Increment<N>; }
const unrelated = "not a number" as number;
