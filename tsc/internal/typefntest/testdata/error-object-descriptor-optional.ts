type D={key:"x";value:1;optional?:false;readonly:false};
typefn F<> { return types.object<[D]>; }
type Result=F<>;
