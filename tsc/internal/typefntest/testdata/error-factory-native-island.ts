export {};
typefn Factory<> {return typefn<T> {return T;};}
typefn Unused<> {return type([Factory<>]);}
