export {};
typefn Bad<> {return types.recursive<typefn<Self> {return types.tuple<{elements:[{value:Self;label:[];mode:"rest"}];readonly:false}>;}>;}
type Result=Bad<>;
