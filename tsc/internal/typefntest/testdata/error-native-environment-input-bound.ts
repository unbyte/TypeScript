export {};
function outer<T>() {typefn Bound<U extends T> {return U;} let result!:Bound<1>;return result;}
type Invalid=ReturnType<typeof outer<2>>;
