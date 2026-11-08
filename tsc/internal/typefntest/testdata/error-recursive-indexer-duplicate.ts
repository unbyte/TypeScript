export {};
typefn Bad<> { return types.recursive<typefn<Self> {return types.object<[],[{key:string;value:Self;readonly:false},{key:string;value:Self;readonly:true}]>;}>; }
type Result=Bad<>;
