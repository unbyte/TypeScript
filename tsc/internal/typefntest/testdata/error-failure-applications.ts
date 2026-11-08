typefn Reject<T> { fail<T>; }
type A = Reject<"bad"> /*one*/;
type B = Reject<"bad"> /*two*/;
