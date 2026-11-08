export {};
typefn Bad<> { return types.recursive<typefn<Self> {type unused = types.object<[{key:"x";value:1;optional:false;readonly:false}],[{key:string;value:Self;readonly:false}]>; return {next:Self};}>; }
type Result=Bad<>;
