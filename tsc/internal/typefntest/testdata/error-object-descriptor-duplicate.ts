type D<K>={key:K;value:1;optional:false;readonly:false};
typefn F<> { return types.object<[D<1>,D<"1">]>; }
type Result=F<>;
