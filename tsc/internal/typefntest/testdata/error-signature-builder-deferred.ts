typefn Make<T> {
 return types.signature<{kind:"call";parameters:[{name:"value";value:T;mode:"required"}];this:[];returns:boolean;abstract:false;predicate:[{kind:"is";target:0;narrowed:[string]}]}>;
}
type Discarded<T> = Make<T> | unknown;
type Result = Discarded<number>;
