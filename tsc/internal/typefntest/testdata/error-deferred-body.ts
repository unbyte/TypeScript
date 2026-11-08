typefn Discard<T> {type unused=T*2;return string;}
function bad<T>(x:any) {const value:Discard<T>=x;}
