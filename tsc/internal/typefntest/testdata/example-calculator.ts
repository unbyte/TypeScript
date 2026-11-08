import type { Calc, CalcBig, ReadInteger, Parsed, A, B, C } from "./calculator";
type Equal<X, Y> = (<T>() => T extends X ? 1 : 2) extends (<T>() => T extends Y ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type Later<S> = Calc<S>;
type Checks = [
 Assert<Equal<Parsed, [123n, 3]>>,
 Assert<Equal<A, 14>>,
 Assert<Equal<B, 20>>,
 Assert<Equal<C, 9007199254740993n>>,
 Assert<Equal<Calc<"20 / 5 / 2">, 2>>,
 Assert<Equal<Calc<"10 - 3 - 2">, 5>>,
 Assert<Equal<Calc<"--2 + -3 * (2 + 1)">, -7>>,
 Assert<Equal<Calc<" \t2\r\n+ 3 ">, 5>>,
 Assert<Equal<Later<"2 * 3 + 4">, 10>>
];
