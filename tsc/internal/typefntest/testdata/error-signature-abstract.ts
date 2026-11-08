abstract class Abstract { abstract value: number; }
typefn Forward<T> { return types.overload<[], reflect.constructs<T>>; }
declare const ctor: Forward<typeof Abstract>;
new ctor();
