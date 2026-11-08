export {};
typefn Factory<>: typefn<_> {return Factory<>;}
typefn Outer<> {type F=Factory<>; return F<1>;}
type Bad=Outer<>;
