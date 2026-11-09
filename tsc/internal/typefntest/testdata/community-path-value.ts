import type { FieldPathValue, FieldArrayPathValue } from "./values";
import type { NativePathValue } from "./reference";
import type { FieldPaths } from "./paths";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
interface Node { value:string; next?:Node }
class Nominal { private brand!: void; nested!: {id:1}; }
type Cases = [
 [{user:{id:1}}, "user.id"], [{user:{id:1}}, "user"],
 [{user?:{id:1}}, "user.id"], [{user:{id:1}|undefined}, "user.id"],
 [{user:{id:1}|null}, "user.id"], [{rows:{id:1}[]}, "rows.0.id"],
 [{rows?:readonly {id:1}[]}, `rows.${number}.id`],
 [readonly [first:{id:1}, second?:{id:2}], "1.id"],
 [readonly [head:{id:1}, ...tail:{id:2}[]], `${number}.id`],
 [{value:never}, never], [{value?:never}, "value"],
 [{value:unknown}, "value"], [{value:any}, "value.anything"],
 [any, "whatever.path"], [any, string], [any, any], [never, never],
 [Record<string,{id:1}>, `${string}.id`], [Record<string,{id:1}>, "entry.id"],
 [Record<string,{id:1}>, string], [{rows:{id:1}[]}, `rows.${number}`],
 [{a:1}|{b:2}, "a"|"b"], [{a?:1}|{a:2}, "a"],
 [{"a.b":{id:1}}, "a.b.id"], [{"1":{id:1}}, "1.id"],
 [Node, "next"], [Nominal, "nested.id"],
 [{date:Date; files:FileList}, "date"|"files"],
 [{fn: (()=>void)&{meta:{id:1}}}, "fn.meta.id"],
 [string[], "0"], [string[], "01"], [string[], `${number}`],
 [{foo:1}, any], [string[], any], [Record<string,{id:1}>, any],
 [{a:{b:1}[]|string[]}, "a.0.b"], [{a?:{b?:{c:1}}}, "a.b.c"]
];
type Compare0 = Assert<Equal<FieldPathValue<Cases[0][0], Cases[0][1]>, NativePathValue<Cases[0][0], Cases[0][1]>>>;
type Compare1 = Assert<Equal<FieldPathValue<Cases[1][0], Cases[1][1]>, NativePathValue<Cases[1][0], Cases[1][1]>>>;
type Compare2 = Assert<Equal<FieldPathValue<Cases[2][0], Cases[2][1]>, NativePathValue<Cases[2][0], Cases[2][1]>>>;
type Compare3 = Assert<Equal<FieldPathValue<Cases[3][0], Cases[3][1]>, NativePathValue<Cases[3][0], Cases[3][1]>>>;
type Compare4 = Assert<Equal<FieldPathValue<Cases[4][0], Cases[4][1]>, NativePathValue<Cases[4][0], Cases[4][1]>>>;
type Compare5 = Assert<Equal<FieldPathValue<Cases[5][0], Cases[5][1]>, NativePathValue<Cases[5][0], Cases[5][1]>>>;
type Compare6 = Assert<Equal<FieldPathValue<Cases[6][0], Cases[6][1]>, NativePathValue<Cases[6][0], Cases[6][1]>>>;
type Compare7 = Assert<Equal<FieldPathValue<Cases[7][0], Cases[7][1]>, NativePathValue<Cases[7][0], Cases[7][1]>>>;
type Compare8 = Assert<Equal<FieldPathValue<Cases[8][0], Cases[8][1]>, NativePathValue<Cases[8][0], Cases[8][1]>>>;
type Compare9 = Assert<Equal<FieldPathValue<Cases[9][0], Cases[9][1]>, NativePathValue<Cases[9][0], Cases[9][1]>>>;
type Compare10 = Assert<Equal<FieldPathValue<Cases[10][0], Cases[10][1]>, NativePathValue<Cases[10][0], Cases[10][1]>>>;
type Compare11 = Assert<Equal<FieldPathValue<Cases[11][0], Cases[11][1]>, NativePathValue<Cases[11][0], Cases[11][1]>>>;
type Compare12 = Assert<Equal<FieldPathValue<Cases[12][0], Cases[12][1]>, NativePathValue<Cases[12][0], Cases[12][1]>>>;
type Compare13 = Assert<Equal<FieldPathValue<Cases[13][0], Cases[13][1]>, NativePathValue<Cases[13][0], Cases[13][1]>>>;
type Compare14 = Assert<Equal<FieldPathValue<Cases[14][0], Cases[14][1]>, NativePathValue<Cases[14][0], Cases[14][1]>>>;
type Compare15 = Assert<Equal<FieldPathValue<Cases[15][0], Cases[15][1]>, NativePathValue<Cases[15][0], Cases[15][1]>>>;
type Compare16 = Assert<Equal<FieldPathValue<Cases[16][0], Cases[16][1]>, NativePathValue<Cases[16][0], Cases[16][1]>>>;
type Compare17 = Assert<Equal<FieldPathValue<Cases[17][0], Cases[17][1]>, NativePathValue<Cases[17][0], Cases[17][1]>>>;
type Compare18 = Assert<Equal<FieldPathValue<Cases[18][0], Cases[18][1]>, NativePathValue<Cases[18][0], Cases[18][1]>>>;
type Compare19 = Assert<Equal<FieldPathValue<Cases[19][0], Cases[19][1]>, NativePathValue<Cases[19][0], Cases[19][1]>>>;
type Compare20 = Assert<Equal<FieldPathValue<Cases[20][0], Cases[20][1]>, NativePathValue<Cases[20][0], Cases[20][1]>>>;
type Compare21 = Assert<Equal<FieldPathValue<Cases[21][0], Cases[21][1]>, NativePathValue<Cases[21][0], Cases[21][1]>>>;
type Compare22 = Assert<Equal<FieldPathValue<Cases[22][0], Cases[22][1]>, NativePathValue<Cases[22][0], Cases[22][1]>>>;
type Compare23 = Assert<Equal<FieldPathValue<Cases[23][0], Cases[23][1]>, NativePathValue<Cases[23][0], Cases[23][1]>>>;
type Compare24 = Assert<Equal<FieldPathValue<Cases[24][0], Cases[24][1]>, NativePathValue<Cases[24][0], Cases[24][1]>>>;
type Compare25 = Assert<Equal<FieldPathValue<Cases[25][0], Cases[25][1]>, NativePathValue<Cases[25][0], Cases[25][1]>>>;
type Compare26 = Assert<Equal<FieldPathValue<Cases[26][0], Cases[26][1]>, NativePathValue<Cases[26][0], Cases[26][1]>>>;
type Compare27 = Assert<Equal<FieldPathValue<Cases[27][0], Cases[27][1]>, NativePathValue<Cases[27][0], Cases[27][1]>>>;
type Compare28 = Assert<Equal<FieldPathValue<Cases[28][0], Cases[28][1]>, NativePathValue<Cases[28][0], Cases[28][1]>>>;
type Compare29 = Assert<Equal<FieldPathValue<Cases[29][0], Cases[29][1]>, NativePathValue<Cases[29][0], Cases[29][1]>>>;
type Compare30 = Assert<Equal<FieldPathValue<Cases[30][0], Cases[30][1]>, NativePathValue<Cases[30][0], Cases[30][1]>>>;
type Compare31 = Assert<Equal<FieldPathValue<Cases[31][0], Cases[31][1]>, NativePathValue<Cases[31][0], Cases[31][1]>>>;
type Compare32 = Assert<Equal<FieldPathValue<Cases[32][0], Cases[32][1]>, NativePathValue<Cases[32][0], Cases[32][1]>>>;
type Compare33 = Assert<Equal<FieldPathValue<Cases[33][0], Cases[33][1]>, NativePathValue<Cases[33][0], Cases[33][1]>>>;
type Compare34 = Assert<Equal<FieldPathValue<Cases[34][0], Cases[34][1]>, NativePathValue<Cases[34][0], Cases[34][1]>>>;
type Compare35 = Assert<Equal<FieldPathValue<Cases[35][0], Cases[35][1]>, NativePathValue<Cases[35][0], Cases[35][1]>>>;
type Compare36 = Assert<Equal<FieldPathValue<Cases[36][0], Cases[36][1]>, NativePathValue<Cases[36][0], Cases[36][1]>>>;
type Optional = Assert<Equal<FieldPathValue<{a?:{b?:{c:1}}}, "a.b.c">, 1|undefined>>;
type Union = Assert<Equal<FieldPathValue<{a:1}|{b:2}, "a"|"b">, 1|2>>;
type Symbolic = Assert<Equal<FieldPathValue<Record<string,{id:1}>, `${string}.id`>, 1>>;
type ArrayValue = Assert<Equal<FieldArrayPathValue<{rows:{id:1}[]}, "rows">, {id:1}[]>>;
type OptionalArray = Assert<Equal<FieldArrayPathValue<{rows?:readonly {id:1}[]}, "rows">, readonly {id:1}[]|undefined>>;
type NestedArray = Assert<Equal<FieldArrayPathValue<{rows:{nested:{id:1}[]}[]}, `rows.${number}.nested`>, {id:1}[]>>;
type Deferred<T extends object, P extends FieldPaths<T>> = FieldPathValue<T,P>;
type Specialized = Assert<Equal<Deferred<{a:{b:1}}, "a.b">, 1>>;
declare function read<T extends object, const P extends FieldPaths<T>>(value:T,path:P):FieldPathValue<T,P>;
const result = read({user:{id:1}}, "user.id");
type Inference = Assert<Equal<typeof result, number>>;
function forward<T extends object, P extends FieldPaths<T>>(value:T,path:P) { return read<T,P>(value,path); }
