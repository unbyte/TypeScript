export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.object<[{key:Self;value:Self;optional:false;readonly:false}]>;}>;}
type Result=Bad<>;
