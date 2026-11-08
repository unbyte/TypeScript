export {};
typefn Bad<> { return types.recursive<typefn<Self> {return types.object<[{key:"x";value:1;optional:false;readonly:false}],[{key:string;value:Self;readonly:false}]>;}>; }
type Result=Bad<>;
