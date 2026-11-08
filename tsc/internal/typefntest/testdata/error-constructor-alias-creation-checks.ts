export {};
typefn Box<T> {return [T];}
typefn Factory<T> {type check=T+1; return Box;}
function outer<T>() {type Invalid=Factory<T>;}
