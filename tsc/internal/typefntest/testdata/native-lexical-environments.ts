export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
function outer<T>() {
 typefn Observe<> {if(T=="yes") {return 1;} return 2;}
 let result!:Observe<>;
 return result;
}
type Yes=Assert<Equal<ReturnType<typeof outer<"yes">>,1>>;
type No=Assert<Equal<ReturnType<typeof outer<"no">>,2>>;
function bounded<T>() {
 typefn Bound<U extends T> {return U;}
 let result!:Bound<1>;
 return result;
}
type Bound=Assert<Equal<ReturnType<typeof bounded<1>>,1>>;
function sibling<T>() {
 typefn Other<> {if(T==1) {return 7;} return 8;}
 typefn Root<> {if(T==1) {return Other<>;} return 0;}
 let result!:Root<>;
 return result;
}
type Sibling=Assert<Equal<ReturnType<typeof sibling<1>>,7>>;
function factory<T>() {
 typefn Make<> {return typefn<> {if(T==1) {return 10;} return 20;};}
 type F=Make<>;
 let result!:F<>;
 return result;
}
type Closure=Assert<Equal<ReturnType<typeof factory<1>>,10>>;
function nativeAlias<T>() {
 type Local<U extends T= T>=[T,U];
 typefn Root<> {if(T==1) {return Local<1>;} return Local;}
 let result!:Root<>;
 return result;
}
type Alias=Assert<Equal<ReturnType<typeof nativeAlias<1>>,[1,1]>>;
type Defaults=Assert<Equal<ReturnType<typeof nativeAlias<2>>,[2,2]>>;
function nativeConstructor<T>() {
 type Local<U extends T>=[T,U];
 typefn Apply<F<_>,U> {return F<U>;}
 typefn Root<> {if(T==1) {return Apply<Local,1>;} return Apply<Local,T>;}
 let result!:Root<>;
 return result;
}
type Constructor=Assert<Equal<ReturnType<typeof nativeConstructor<1>>,[1,1]>>;
class Container<T> {
 method() {
  typefn Choose<> {if(T==1) {return 123;} return 456;}
  let result!:Choose<>;
  return result;
 }
}
type Method=Assert<Equal<ReturnType<Container<1>["method"]>,123>>;
function nested<T>() {
 return function<U>() {
  typefn Choose<> {if(T==1) {if(U==2) {return 12;} return 11;} return 0;}
  let result!:Choose<>;
  return result;
 };
}
const selected=nested<1>();
const other=nested<2>();
type TwiceMapped=Assert<Equal<ReturnType<typeof selected<2>>,12>>;
type Independent=Assert<Equal<ReturnType<typeof other<2>>,0>>;
// Force specialization before these declarations receive their ordinary source
// checks. Cached native declaration Types must stay in their defining environment.
type LazyOne=Assert<Equal<ReturnType<typeof lazy<1>>,[3,1]>>;
type LazyTwo=Assert<Equal<ReturnType<typeof lazy<2>>,[4,2]>>;
function lazy<T>() {
 typefn Choose<U> {if(T==1) {return [U,1];} return [U,2];}
 type Native<U>=Choose<U>;
 typefn Root<> {if(T==1) {return Native<3>;} return Native<4>;}
 let result!:Root<>;
 return result;
}
type AliasOne=Assert<Equal<ReturnType<typeof lazyAlias<1>>,[3,1]>>;
type AliasTwo=Assert<Equal<ReturnType<typeof lazyAlias<2>>,[3,2]>>;
function lazyAlias<T>() {
 typefn Make<> {return typefn<U> {if(T==1) {return [U,1];} return [U,2];};}
 type Alias=Make<>;
 typefn Root<> {if(T==1) {return Alias<3>;} return Alias<3>;}
 let result!:Root<>;
 return result;
}
