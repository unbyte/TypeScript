export {};
function outer<T>() {type Local<U extends T>=[T,U];typefn F<> {if(T==2) {return type(Local<1>);} return 0;} let result!:F<>;return result;}
type Invalid=ReturnType<typeof outer<2>>;
