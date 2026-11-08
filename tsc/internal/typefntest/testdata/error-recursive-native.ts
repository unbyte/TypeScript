export {};
typefn Bad<> {return types.recursive<typefn<Self> {return Partial<Self>;}>;}
type Result=Bad<>;
