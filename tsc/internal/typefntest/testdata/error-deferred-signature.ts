typefn Discard<T> {type unused=T*2;return string;}
declare const partial:<T>(x:T)=>Discard<T>;
const total:<T>(x:T)=>string=partial;
