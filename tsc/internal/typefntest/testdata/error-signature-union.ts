type Fn = (() => 1) | (() => 2);
typefn Bad<> { return reflect.calls<Fn>; }
type Result = Bad<>;
