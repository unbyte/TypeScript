export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Reflect<T> { return reflect.tuple<T>; }
typefn Rebuild<T> { return types.tuple<reflect.tuple<T>>; }
const holes = [,1,undefined] as const;
type Hole = Assert<Equal<Reflect<typeof holes>['elements'][0]['value'],never>>;
type Following = Assert<Equal<Reflect<typeof holes>['elements'][1]['value'],1>>;
type Explicit = Assert<Equal<Reflect<typeof holes>['elements'][2]['value'],undefined>>;
type Mode = Assert<Equal<Reflect<typeof holes>['elements'][0]['mode'],'optional'>>;
type Roundtrip = Assert<Equal<Rebuild<typeof holes>,typeof holes>>;
