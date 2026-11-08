export {};
typefn Bad<> {return types.recursive<typefn<Self> {type F=typefn<U> {return {next:Self;value:U};};return type(F<1>);}>;}
type Result=Bad<>;
