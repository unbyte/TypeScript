export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.constructor<[],Self,Self>;}>;}
type Result=Bad<>;
