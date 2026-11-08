export {};
typefn Bad<> {return types.recursive<typefn<Self>{return types.signature<{kind:"call";parameters:[{name:"x";value:Self;mode:"rest"}];this:[];returns:Self;abstract:false;predicate:[]}>;}>;}
type Result=Bad<>;
