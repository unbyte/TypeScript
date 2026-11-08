export {};
type Assert<T extends true>=T;
type Equal<A,B>=(<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
function length<T>() {typefn F<> {return text.length<T>;} let result!:F<>;return result;}
type Length=Assert<Equal<ReturnType<typeof length<"abc">>,3>>;
function keyed<T>() {typefn F<> {return {[T]:1};} let result!:F<>;return result;}
type Key=Assert<Equal<ReturnType<typeof keyed<"a">>,{a:1}>>;
function sliced<T>() {typefn F<> {return tuple.slice<T,1,2>;} let result!:F<>;return result;}
type Slice=Assert<Equal<ReturnType<typeof sliced<[1,2,3]>>,[2]>>;
function build<T>() {typefn F<> {type value=1;type saved={[T]:value};value=2;return saved;} let result!:F<>;return result;}
type Snapshot=Assert<Equal<ReturnType<typeof build<"a">>,{a:1}>>;
