class Hidden { private constructor() {} }
typefn Bad<T> { return types.overload<[], [T]>; }
type Result = Bad<typeof Hidden>;
