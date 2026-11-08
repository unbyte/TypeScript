typefn Ctor<A> { return types.constructor<[], {value:number}, A>; }
declare const ctor: Ctor<true>;
new ctor();
