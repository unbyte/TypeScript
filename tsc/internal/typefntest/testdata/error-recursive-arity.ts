export {};
typefn Bad<> {return types.recursive<typefn<A,B> {return A;}>;}
