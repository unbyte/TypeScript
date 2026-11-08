typefn F<T> {return types.object<[],[{key:string;value:T;readonly:true}]>; }
type Later<T>=F<T>;
declare let value:Later<number>;
value["x"]=1;
