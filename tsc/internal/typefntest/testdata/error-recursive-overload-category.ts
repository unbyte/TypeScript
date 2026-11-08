export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.object<[],[],[types.constructor<[],Self>]>;}>;}
type Result=Bad<>;
