export {};
typefn Box<T> {return [T];}
typefn Pair<A,B> {return [A,B];}
typefn Unused<> {type F=Box; F=Pair; return 0;}
