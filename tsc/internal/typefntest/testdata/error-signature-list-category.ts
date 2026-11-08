type Fn = () => void;
typefn Bad<> { return types.overload<[], [Fn]>; }
type Result = Bad<>;
