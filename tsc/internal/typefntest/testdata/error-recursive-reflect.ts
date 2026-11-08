export {};
typefn Bad<> {return types.recursive<typefn<Self> {return reflect.kind<Self>;}>;}
type Result=Bad<>;
