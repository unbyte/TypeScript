type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Pair<T> { if (T == 1) { return [42, "yes"]; } return [0, "no"]; }
typefn First<T> { return T[0]; }
typefn Select<T> { return Pair<T>[1]; }
type Later<T> = Select<T>;
type LaterFirst<T> = First<T>;
type Checks = [Assert<Equal<Later<1>, "yes">>, Assert<Equal<Later<2>, "no">>, Assert<Equal<LaterFirst<[7]>, 7>>];
