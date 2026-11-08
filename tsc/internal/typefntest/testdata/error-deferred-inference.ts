typefn Twice<N> { return N * 2; }
declare function double<N>(value: N): Twice<N>;
const first = double("x" as const);
const second = double("x" as const);
const answer: 42 = double(21 as const);
