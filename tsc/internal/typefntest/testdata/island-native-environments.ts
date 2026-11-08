export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
function direct<T>() {typefn F<> {if(type(T)==1) {return 10;} return 20;} let result!:F<>;return result;}
type Direct=Assert<Equal<ReturnType<typeof direct<1>>,10>>;
function alias<T>() {
 type Local<U=T>=[T,U];
 typefn F<> {if(T==1) {return type(Local<2>);} return type(Local);}
 let result!:F<>;return result;
}
type Alias=Assert<Equal<ReturnType<typeof alias<1>>,[1,2]>>;
type Default=Assert<Equal<ReturnType<typeof alias<3>>,[3,3]>>;
function helper<T>() {
 typefn Other<> {if(T==1) {return 10;} return 20;}
 typefn F<> {if(T==1) {return type(Other<>);} return type(Other<>);}
 let result!:F<>;return result;
}
type Helper=Assert<Equal<ReturnType<typeof helper<1>>,10>>;
function constructorAlias<T>() {
 typefn Make<> {return typefn<U> {if(T==1) {return [U,1];} return [U,2];};}
 type Alias=Make<>;
 typefn F<> {if(T==1) {return type(Alias<3>);} return type(Alias<4>);}
 let result!:F<>;return result;
}
type Closure=Assert<Equal<ReturnType<typeof constructorAlias<1>>,[3,1]>>;
type ClosureTwo=Assert<Equal<ReturnType<typeof constructorAlias<2>>,[4,2]>>;
function values<T>(input:{value:T}) {
 typefn F<> {if(T==1) {return type(typeof input.value);} return type(typeof input);}
 let result!:F<>;return result;
}
type Value=Assert<Equal<ReturnType<typeof values<1>>,1>>;
type ObjectValue=Assert<Equal<ReturnType<typeof values<2>>,{value:2}>>;
function signatures<T>() {
 function identity<U extends T>(value:U):U {return value;}
 typefn F<> {if(T==1) {return type(typeof identity<1>);} return type(typeof identity<T>);}
 let result!:F<>;return result;
}
type Signature=Assert<Equal<ReturnType<typeof signatures<1>>,(value:1)=>1>>;
type SignatureTwo=Assert<Equal<ReturnType<typeof signatures<2>>,(value:2)=>2>>;
typefn LocalBinders<> {return type(<T>(value:T)=>typeof value);}
type OwnBinders=Assert<Equal<LocalBinders<>,<T>(value:T)=>T>>;
type MixedNames=Assert<Equal<ReturnType<typeof merged<1>>,[[2],1]>>;
function merged<T>(Shared:{value:T}) {
 typefn Shared<U> {return [U];}
 typefn F<> {if(T==1) {return type([Shared<2>,typeof Shared.value]);} return 0;}
 let result!:F<>;return result;
}
// Query before source checking to exercise declaration cache isolation.
type LazyOne=Assert<Equal<ReturnType<typeof lazy<1>>,[1,1]>>;
type LazyTwo=Assert<Equal<ReturnType<typeof lazy<2>>,[2,2]>>;
function lazy<T>() {
 typefn Choose<U> {if(T==1) {return [U,1];} return [U,2];}
 type Native<U>=Choose<U>;
 typefn F<> {if(T==1) {return type(Native<1>);} return type(Native<2>);}
 let result!:F<>;return result;
}
type ValueOne=Assert<Equal<ReturnType<typeof lazyValue<1>>,1>>;
type ValueTwo=Assert<Equal<ReturnType<typeof lazyValue<2>>,2>>;
function lazyValue<T>() {
 typefn Shape<> {if(T==1) {return {value:1};} return {value:2};}
 let input!:Shape<>;
 typefn F<> {if(T==1) {return type(typeof input.value);} return type(typeof input.value);}
 let result!:F<>;return result;
}
