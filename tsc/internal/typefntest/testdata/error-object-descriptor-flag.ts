typefn F<O> { return types.object<[{key:"x";value:1;optional:O;readonly:false}]>; }
type Result=F<boolean>;
