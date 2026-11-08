export {};
typefn Bad<> {return types.recursive<typefn<Self>{return types.signature<{kind:"call";parameters:[{name:"x";value:unknown;mode:"required"}];this:[];returns:Self;abstract:false;predicate:[{kind:"is";target:0;narrowed:[Self]}]}>;}>;}
type Result=Bad<>;
