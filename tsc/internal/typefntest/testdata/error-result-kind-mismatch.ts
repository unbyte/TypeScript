export {};
typefn Box<T> {return T;}
typefn Mixed<T> {if(T) {return 1;} return Box;}
typefn Annotated<>: type {return Box;}
