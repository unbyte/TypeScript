export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.object<[{key:0;value:Self;optional:false;readonly:false},{key:"0";value:Self;optional:false;readonly:false}]>;}>;}
type Result=Bad<>;
