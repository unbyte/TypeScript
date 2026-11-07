type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Box<T> { return {value:T}; }
function unwrap<T>(x:Box<T>):T { return x.value; }
const value=unwrap({value:"hello" as const});
type Inferred=Assert<Equal<typeof value,"hello">>;
typefn Twice<T> { return T*2; }
declare function double<T>(x:T):Twice<T>;
const result=double(21 as const);
type Computed=Assert<Equal<typeof result,42>>;
function identity<T>(x:Twice<T>):Twice<T> {return x;}
typefn Discard<T> {type unused=T*2;return string;}
function guardedIdentity<T>(x:Discard<T>):Discard<T> {return x;}
