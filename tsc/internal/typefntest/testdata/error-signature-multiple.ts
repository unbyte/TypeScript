interface Fn { (): 1; (x: string): 2; }
typefn Bad<> { return types.overload<[Fn]>; }
type Result = Bad<>;
