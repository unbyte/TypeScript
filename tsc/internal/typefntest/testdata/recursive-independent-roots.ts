export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Independent<Count,Payload> {return types.recursiveGroup<Count,typefn<R> {
 type result=[];
 type index=0;
 while(index<tuple.length<R>) {
  result=[...result,{index:index;payload:Payload;next:R[index]}];
  index=index+1;
 }
 return result;
}>;}
type Group<T>=Independent<256,T>;
type Size=Assert<Equal<Group<1>['length'],256>>;
type First=Assert<Equal<Group<1>[0]['next']['index'],0>>;
type Last=Assert<Equal<Group<1>[255]['next']['index'],255>>;
type Payload=Assert<Equal<Group<2>[255]['next']['payload'],2>>;
