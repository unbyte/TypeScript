typefn Ctor<A> { return types.constructor<[], {}, A>; }
type Result = Ctor<0>;
