export {};
typefn Bad<> {return types.recursive<typefn<Self>{return types.signature<{kind:"call";parameters:[{name:"x";value:number;mode:"required"}];this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:0;narrowed:[Self]}]}>;}>;}
type Result=Bad<>;
