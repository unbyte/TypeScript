export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.overload<[Self]>;}>;}
type Result=Bad<>;
