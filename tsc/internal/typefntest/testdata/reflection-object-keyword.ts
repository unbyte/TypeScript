export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Public<T> { return [reflect.kind<T>,reflect.properties<T>,reflect.indexers<T>,reflect.calls<T>,reflect.constructs<T>]; }
type EmptyView = Assert<Equal<Public<object>,["object",[],[],[],[]]>>;
type EmptyStructuralView = Assert<Equal<Public<{}>,["object",[],[],[],[]]>>;
type Deferred<T> = Public<T>;
type Replay = Assert<Equal<Deferred<object>,["object",[],[],[],[]]>>;
