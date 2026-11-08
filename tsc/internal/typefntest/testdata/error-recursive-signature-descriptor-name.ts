export {};
typefn Bad<> {return types.recursive<typefn<Self>{return types.signature<{kind:"call";parameters:[{name:Self;value:Self;mode:"required"}];this:[];returns:void;abstract:false;predicate:[]}>;}>;}
type Result=Bad<>;
