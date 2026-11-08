typefn F<K> { return types.object<[{key:K;value:1;optional:false;readonly:false}]>; }
type Later<K>=F<K>;
type Result=Later<"x"|"y">;
