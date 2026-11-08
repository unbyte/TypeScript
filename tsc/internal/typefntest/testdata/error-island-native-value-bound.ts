export {};
function outer<T>() {function identity<U extends T>(value:U):U {return value;} typefn F<> {if(T==2) {return type(typeof identity<1>);} return 0;} let result!:F<>;return result;}
type Invalid=ReturnType<typeof outer<2>>;
