export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Classify<T> {
 type built = types.recursive<typefn<Self> {
  return types.overload<[types.function<[Self],Self>,types.function<[],T>]>;
 }>;
 type signature = reflect.calls<built>[1];
 if (signature is type(()=>number)) { return true; }
 return false;
}
type Deferred<T>=Classify<T>;
type A=Assert<Equal<Deferred<1>,true>>;
type B=Assert<Equal<Deferred<"x">,false>>;
type C=Assert<Equal<Deferred<2>,true>>;
