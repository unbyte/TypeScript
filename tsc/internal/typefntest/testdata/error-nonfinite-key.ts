typefn Static<> { return { 1e999: string }; }
typefn Computed<N> { return { [N]: string }; }
type A = Static<>;
type B = Computed<1e999>;
