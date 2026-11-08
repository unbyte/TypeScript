export {};
typefn Bad<> { return types.recursive<typefn<Self> {return types.object<[],[{key:Self;value:Self;readonly:false}]>;}>; }
type Result=Bad<>;
