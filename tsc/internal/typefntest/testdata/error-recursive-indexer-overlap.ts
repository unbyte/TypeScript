export {};
typefn Bad<> { return types.recursive<typefn<Self> {return types.object<[],[{key:string;value:Self;readonly:false},{key:number;value:1;readonly:false}]>;}>; }
type Result=Bad<>;
