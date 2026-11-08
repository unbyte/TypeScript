typefn Project<T> {
 type fields=[];
 for(type field of reflect.properties<T>) {
  fields=[...fields,{key:field["key"];value:field["value"];optional:field["optional"];readonly:true}];
 }
 return types.object<fields>;
}
declare let result:Project<{a:string}>;
result.a="changed";
