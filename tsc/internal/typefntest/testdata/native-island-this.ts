export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
class Box<T> {
 value!:T;
 identity<U>(value:U) {return value;}
 signature<U>() {
  typefn Read<X> {return type(typeof this.identity<X>);}
  let result!:Read<U>; return result;
 }
 callback() {
  typefn Read<> {return type((value:this)=>this);}
  let result!:Read<>; return result;
 }

 self() {
  typefn Read<> {return type(this);}
  type Result=Read<>;
  let result!:Result; return result;
 }
 query() {
  typefn Read<> {return type(typeof this);}
  type Result=Read<>;
  let result!:Result; return result;
 }
 member() {
  typefn Read<> {return type(typeof this.value);}
  type Result=Read<>;
  let result!:Result; return result;
 }
}
class Derived extends Box<3> {extra!:true;}
declare const a:Box<1>;
declare const b:Box<2>;
declare const d:Derived;
const selfA=a.self(); const selfB=b.self(); const selfD=d.self();
type A=Assert<Equal<typeof selfA.value,1>>;
type B=Assert<Equal<typeof selfB.value,2>>;
type D=Assert<Equal<typeof selfD.extra,true>>;
const queryA=a.query(); const queryB=b.query(); const queryD=d.query();
type QA=Assert<Equal<typeof queryA.value,1>>;
type QB=Assert<Equal<typeof queryB.value,2>>;
type QD=Assert<Equal<typeof queryD.extra,true>>;
const memberA=a.member(); const memberB=b.member();
type MA=Assert<Equal<typeof memberA,1>>;
type MB=Assert<Equal<typeof memberB,2>>;
function receiver<T>(this:{value:T}) {
 typefn Read<> {return type(typeof this.value);}
 type Result=Read<>;
 let result!:Result; return result;
}
const bound=receiver<"bound">;
const result=bound.call({value:"bound"});
type R=Assert<Equal<typeof result,"bound">>;

class Static {
 static tag=42 as const;
 static read() {
  typefn Read<> {return type(typeof this.tag);}
  let result!:Read<>; return result;
 }
}
type StaticThis=Assert<Equal<ReturnType<typeof Static.read>,42>>;

const signature=d.signature<"literal">();
type QueryArguments=Assert<Equal<Parameters<typeof signature>,[value:"literal"]>>;
type QueryResult=Assert<Equal<ReturnType<typeof signature>,"literal">>;
const callback=d.callback();
type Callback=Assert<Equal<ReturnType<typeof callback>,Derived>>;
typefn Predicate<> {return type(()=>this is {ready:true});}
declare const receiverWithPredicate:{check:Predicate<>};
if(receiverWithPredicate.check()){const ready:true=receiverWithPredicate.ready;}
