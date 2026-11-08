export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Make<T> {return types.recursive<typefn<Self> {
 type Inner=typefn<U> {return {next:Self;value:U};};
 return Inner<T>;
}>;}
typefn Repeat<> {
 type count=0;
 while(count<100) {type unused=Make<1>;count=count+1;}
 return count;
}
type Repeated=Assert<Equal<Repeat<>,100>>;
