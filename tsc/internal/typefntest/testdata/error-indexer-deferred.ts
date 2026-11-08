typefn F<T> {return types.object<[{key:"x";value:T;optional:false;readonly:false}],[{key:string;value:string;readonly:false}]>; }
type Later<T>=F<T>|unknown;
type Result=Later<1>;
