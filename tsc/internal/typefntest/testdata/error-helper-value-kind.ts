export {};
typefn Box<T> {return [T];}
typefn Unused<> {type F=Box; F=123; return 0;}
typefn Other<> {type T=123; T=Box; return 0;}
