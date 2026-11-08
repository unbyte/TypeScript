export {};
typefn Box<T> {return [T];}
typefn Bounded<T extends string> {return T;}
typefn Caller<> {type F=Box; F=Bounded; return F<123>;}
type Bad=Caller<>;
