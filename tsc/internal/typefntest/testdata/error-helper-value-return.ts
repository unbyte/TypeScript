export {};
typefn Box<T> {return [T];}
typefn Unused<> {type F=Box; return F;}
