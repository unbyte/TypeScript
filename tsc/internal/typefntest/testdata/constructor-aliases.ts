export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Box<T> {return [T];}
typefn Compose<F<_>,G<_>> {return typefn<T> {return F<G<T>>;};}
type AsyncArray=Compose<Promise,Array>;
type Names=Assert<Equal<AsyncArray<string>,Promise<string[]>>>;
type Named=Box;
type Synonym=(Named);
type NamedResult=Assert<Equal<Synonym<1>,[1]>>;
typefn Apply<F<_>,T> {return F<T>;}
type Passed=Assert<Equal<Apply<AsyncArray,1>,Promise<1[]>>>;
typefn DSL<T> {type F=AsyncArray; return [F<T>,AsyncArray<T>,Apply<Named,T>];}
type DslResult=Assert<Equal<DSL<1>,[Promise<1[]>,Promise<1[]>,[1]]>>;
typefn Island<T> {return type(AsyncArray<T>);}
type IslandResult=Assert<Equal<Island<1>,Promise<1[]>>>;
typefn Capture<T> {return typefn<U> {return [T,U];};}
function outer<T>() {
 type F=Capture<T>;
 let result!:F<1>;
 return result;
}
type Captured=Assert<Equal<ReturnType<typeof outer<2>>,[2,1]>>;
typefn Zero<> {return 123;}
type Nullary=Zero;
type Empty=Assert<Equal<Nullary<>,123>>;
typefn ReturnNative<>: typefn<_> {return Array;}
type Native=ReturnNative<>;
type NativeResult=Assert<Equal<Native<1>,1[]>>;
typefn NeedString<T extends string> {return [T];}
type Bounded=NeedString;
type BoundResult=Assert<Equal<Bounded<"ok">,["ok"]>>;
typefn Scalar<T> {return T+1;}
type ScalarAlias=Scalar;
type Pending<T>=ScalarAlias<T>;
type Specialization=Assert<Equal<Pending<1>,2>>;
// Alias kinds exist before compiling the referenced helper's recursive body.
type Forward=Middle;
type Middle=Countdown;
typefn Countdown<T>: type {if(T==0) {return T;} return Forward<T-1>;}
type ForwardResult=Assert<Equal<Forward<3>,0>>;
