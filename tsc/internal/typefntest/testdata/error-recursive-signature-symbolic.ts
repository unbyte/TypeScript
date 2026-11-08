export {};
typefn Build<T> {return types.recursive<typefn<Self>{
 type unused=types.signature<{kind:"call";parameters:[{name:"x";value:T;mode:"required"}];this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:0;narrowed:[Self]}]}>;
 return 1;
}>;}
type Generic<T>=Build<T>;
type Good=Generic<number>;
type Bad=Generic<string>;
