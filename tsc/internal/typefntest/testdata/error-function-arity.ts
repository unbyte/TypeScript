typefn Fn<P> { return types.function<P, string>; }
declare const fn: Fn<[string, number]>;
fn("only");
