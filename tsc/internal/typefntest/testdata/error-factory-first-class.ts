export {};
typefn Factory<> {return typefn<T> {return T;};}
typefn Unused<> {type F=Factory; return 0;}
