export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Copy<T> { return [...T]; }
typefn Join<A,B> { return [...A,...B]; }
typefn Surround<T> { return [0,...T,1]; }
type Optional = [name:string,count?:number];
type NativeOptional = [...Optional];
type OptionalCheck = Assert<Equal<Copy<Optional>,NativeOptional>>;
type Readonly = readonly [name:string,count?:number];
type NativeReadonly = [...Readonly];
type ReadonlyCheck = Assert<Equal<Copy<Readonly>,NativeReadonly>>;
type Rest = [name:string,...counts:number[]];
type NativeRest = [0,...Rest,1];
type RestCheck = Assert<Equal<Surround<Rest>,NativeRest>>;
type NativeArray = [0,...number[],1];
type ArrayCheck = Assert<Equal<Surround<number[]>,NativeArray>>;
type OptionalThenRequired = [...Optional, true];
type Mixed = Assert<Equal<Join<Optional,[true]>,OptionalThenRequired>>;
type Choice = [1]|[2,3];
type NativeChoice = [...Choice,"end"];
type ChoiceCheck = Assert<Equal<Join<Choice,["end"]>,NativeChoice>>;
type Symbolic<T> = Copy<[value:T,count?:number]>;
type Leaves = Assert<Equal<Symbolic<"x">,[value:"x",count?:number]>>;
type UnknownShape<T> = Surround<T>;
type Replay = Assert<Equal<UnknownShape<[2,3]>,[0,2,3,1]>>;
type RestThenOptional = [0,...number[],...Optional];
type RestOptional = Assert<Equal<Join<[0,...number[]],Optional>,RestThenOptional>>;
