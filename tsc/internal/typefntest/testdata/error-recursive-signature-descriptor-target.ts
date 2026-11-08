export {};
typefn Bad<> {return types.recursive<typefn<Self>{return types.signature<{kind:"call";parameters:[{name:"x";value:unknown;mode:"required"}];this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:Self;narrowed:[Self]}]}>;}>;}
type Result=Bad<>;
