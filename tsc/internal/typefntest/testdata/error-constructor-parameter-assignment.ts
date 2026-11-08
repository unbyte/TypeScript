export {};
typefn Box<T> {return [T];}
typefn Bad<F<_>> {F=Box; return 1;}
