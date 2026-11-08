type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Snapshot<T> {
 type subject = T;
 type result = [];
 for (type item of subject) {
  subject = [];
  result = [...result, item];
 }
 return result;
}
typefn Nested<> {
 type result = [];
 for (type outer of [1, 2]) {
  for (type inner of [1, 2, 3, 4]) {
   if (inner == 2) { continue; }
   if (inner == 4) { break; }
   result = [...result, [outer, inner]];
  }
 }
 return result;
}
typefn Count<T> {
 type count = 0;
 for (type member in T) { count = count + 1; }
 return count;
}
type Symbolic<T> = Snapshot<[T, {value: T}]>;
type Checks = [
 Assert<Equal<Snapshot<readonly [1, 2, 3]>, [1, 2, 3]>>,
 Assert<Equal<Snapshot<[]>, []>>,
 Assert<Equal<Nested<>, [[1, 1], [1, 3], [2, 1], [2, 3]]>>,
 Assert<Equal<Count<never>, 0>>,
 Assert<Equal<Count<any>, 1>>,
 Assert<Equal<Count<unknown>, 1>>,
 Assert<Equal<Count<boolean>, 2>>,
 Assert<Equal<Symbolic<42>, [42, {value: 42}]>>
];
