export {};
typefn Build<T> { return types.recursive<typefn<Self> {
 type unused = types.object<[{key:"x";value:T;optional:false;readonly:false}],[{key:string;value:Self|number;readonly:false}]>;
 return {next:Self};
}>; }
type Generic<T> = Build<T>;
type Good = Generic<1>;
type Bad = Generic<"wrong">;
