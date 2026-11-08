export {};
typefn Box<T> {return [T];}
typefn Choose<T> {if(T==1) {return Box;} return Box;}
function outer<T>() {type Invalid=Choose<T>;}
