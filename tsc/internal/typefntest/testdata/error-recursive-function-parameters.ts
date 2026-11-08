export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.function<Self,1>;}>;}
type Result=Bad<>;
