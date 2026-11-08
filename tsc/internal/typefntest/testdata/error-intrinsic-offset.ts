typefn Bad<>: type { failAt<"bad", -1>; }
type Result = Bad<>;
