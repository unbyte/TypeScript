export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Branch<A,B> {return types.recursive<typefn<Outer> {
 type child=types.recursive<typefn<Inner> {return {payload:B;next:Inner};}>;
 return {payload:A;next:Outer;child:child};
}>;}
type Generic<A,B>=Branch<A,B>;
type Root=Assert<Equal<Generic<1,2>['next']['payload'],1>>;
type Child=Assert<Equal<Generic<1,2>['next']['child']['next']['payload'],2>>;
type Other=Assert<Equal<Generic<3,4>['child']['next']['payload'],4>>;
typefn EmptyChild<T> {return types.recursive<typefn<Outer> {
 type child=types.recursive<typefn<Inner> {return {next:Inner};}>;
 return {value:T;next:Outer;child:child};
}>;}
typefn Kind<T> {return reflect.kind<T>;}
type NoChildDependency<T>=Assert<Equal<Kind<EmptyChild<T>['child']>,"object">>;
type EmptyCycle=Assert<Equal<EmptyChild<1>['child']['next'],EmptyChild<2>['child']>>;
