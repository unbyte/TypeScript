export {};
typefn Bad<> { return types.recursive<typefn<Self> {type unused = types.object<[{key:"x";value:2;optional:false;readonly:false}],[{key:string;value:Self;readonly:false}]>; return 1;}>; }
type Result=Bad<>;
