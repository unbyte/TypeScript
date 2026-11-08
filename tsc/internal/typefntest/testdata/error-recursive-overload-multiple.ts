export {};
typefn Bad<> {return types.recursive<typefn<Self> {type many=types.overload<[types.function<[],Self>,types.function<[Self],void>]>; return types.overload<[many]>;}>;}
type Result=Bad<>;
