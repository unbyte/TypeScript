import type { FieldArrayPaths, FieldArrayPathsAt } from "./paths";
import type { NativeArrayPaths, NativeArrayPathsAt } from "./reference";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
interface Node { children: Node[]; data: {id: number}[] }
interface Left { rights: Right[] }
interface Right { left: Left; rows: {id: number}[] }
declare const symbol: unique symbol;
class Nominal { private brand!: void; rows!: {id: number}[]; }
type Cases = [
 {}, { rows: {id: number}[] }, { strings: string[]; dates: Date[]; files: File[] },
 { rows?: readonly {id: number}[] }, { rows: ({id: number} | string)[] },
 { rows: never[] }, { rows: unknown[] }, { rows: any[] }, { value: any },
 { value: unknown }, { value: never }, { value?: never },
 Node, Left, readonly [first: {id: number}[], second?: {key: string}[]],
 readonly [head: {id: number}[], ...tail: {key: string}[][]],
 { rows: readonly [string, {id: number}] }, { rows: [] },
 { rows: {id: number}[][] }, {rows: {nested: {id: number}[]}[]},
 Record<string, {id: number}[]>, { [key: number]: {id: number}[] },
 {0: {id: number}[]; "1": {id: number}[]; [symbol]: {id: number}[]},
 Nominal, { a: {id: number}[] } | { b: {key: string}[] },
 { rows: {id: number}[] | {id: number} },
 { fn: (() => void) & { rows: {id: number}[] } },
 { "a.b": { rows: {id: number}[] } }, any, never, never[], [],
 { rows: (Date | File | null)[] }, { rows: readonly (unknown | undefined)[] },
 { rows: readonly {id: number}[] | undefined }
];
type Compare0 = Assert<Equal<FieldArrayPaths<Cases[0]>, NativeArrayPaths<Cases[0]>>>;
type Compare1 = Assert<Equal<FieldArrayPaths<Cases[1]>, NativeArrayPaths<Cases[1]>>>;
type Compare2 = Assert<Equal<FieldArrayPaths<Cases[2]>, NativeArrayPaths<Cases[2]>>>;
type Compare3 = Assert<Equal<FieldArrayPaths<Cases[3]>, NativeArrayPaths<Cases[3]>>>;
type Compare4 = Assert<Equal<FieldArrayPaths<Cases[4]>, NativeArrayPaths<Cases[4]>>>;
type Compare5 = Assert<Equal<FieldArrayPaths<Cases[5]>, NativeArrayPaths<Cases[5]>>>;
type Compare6 = Assert<Equal<FieldArrayPaths<Cases[6]>, NativeArrayPaths<Cases[6]>>>;
type Compare7 = Assert<Equal<FieldArrayPaths<Cases[7]>, NativeArrayPaths<Cases[7]>>>;
type Compare8 = Assert<Equal<FieldArrayPaths<Cases[8]>, NativeArrayPaths<Cases[8]>>>;
type Compare9 = Assert<Equal<FieldArrayPaths<Cases[9]>, NativeArrayPaths<Cases[9]>>>;
type Compare10 = Assert<Equal<FieldArrayPaths<Cases[10]>, NativeArrayPaths<Cases[10]>>>;
type Compare11 = Assert<Equal<FieldArrayPaths<Cases[11]>, NativeArrayPaths<Cases[11]>>>;
type Compare12 = Assert<Equal<FieldArrayPaths<Cases[12]>, NativeArrayPaths<Cases[12]>>>;
type Compare13 = Assert<Equal<FieldArrayPaths<Cases[13]>, NativeArrayPaths<Cases[13]>>>;
type Compare14 = Assert<Equal<FieldArrayPaths<Cases[14]>, NativeArrayPaths<Cases[14]>>>;
type Compare15 = Assert<Equal<FieldArrayPaths<Cases[15]>, NativeArrayPaths<Cases[15]>>>;
type Compare16 = Assert<Equal<FieldArrayPaths<Cases[16]>, NativeArrayPaths<Cases[16]>>>;
type Compare17 = Assert<Equal<FieldArrayPaths<Cases[17]>, NativeArrayPaths<Cases[17]>>>;
type Compare18 = Assert<Equal<FieldArrayPaths<Cases[18]>, NativeArrayPaths<Cases[18]>>>;
type Compare19 = Assert<Equal<FieldArrayPaths<Cases[19]>, NativeArrayPaths<Cases[19]>>>;
type Compare20 = Assert<Equal<FieldArrayPaths<Cases[20]>, NativeArrayPaths<Cases[20]>>>;
type Compare21 = Assert<Equal<FieldArrayPaths<Cases[21]>, NativeArrayPaths<Cases[21]>>>;
type Compare22 = Assert<Equal<FieldArrayPaths<Cases[22]>, NativeArrayPaths<Cases[22]>>>;
type Compare23 = Assert<Equal<FieldArrayPaths<Cases[23]>, NativeArrayPaths<Cases[23]>>>;
type Compare24 = Assert<Equal<FieldArrayPaths<Cases[24]>, NativeArrayPaths<Cases[24]>>>;
type Compare25 = Assert<Equal<FieldArrayPaths<Cases[25]>, NativeArrayPaths<Cases[25]>>>;
type Compare26 = Assert<Equal<FieldArrayPaths<Cases[26]>, NativeArrayPaths<Cases[26]>>>;
type Compare27 = Assert<Equal<FieldArrayPaths<Cases[27]>, NativeArrayPaths<Cases[27]>>>;
type Compare28 = Assert<Equal<FieldArrayPaths<Cases[28]>, NativeArrayPaths<Cases[28]>>>;
type Compare29 = Assert<Equal<FieldArrayPaths<Cases[29]>, NativeArrayPaths<Cases[29]>>>;
type Compare30 = Assert<Equal<FieldArrayPaths<Cases[30]>, NativeArrayPaths<Cases[30]>>>;
type Compare31 = Assert<Equal<FieldArrayPaths<Cases[31]>, NativeArrayPaths<Cases[31]>>>;
type Compare32 = Assert<Equal<FieldArrayPaths<Cases[32]>, NativeArrayPaths<Cases[32]>>>;
type Compare33 = Assert<Equal<FieldArrayPaths<Cases[33]>, NativeArrayPaths<Cases[33]>>>;
type Compare34 = Assert<Equal<FieldArrayPaths<Cases[34]>, NativeArrayPaths<Cases[34]>>>;
type Fixed = Assert<Equal<FieldArrayPaths<{rows:{items:{id:number}[]}[]; strings:string[]}>, "rows" | `rows.${number}.items`>>;
type Recursive = Assert<Equal<FieldArrayPaths<Node>, "children" | "data">>;
type Shallow = Assert<Equal<FieldArrayPathsAt<{a:{rows:{id:number}[]}; root:{id:number}[]}, 0>, "root">>;
type TwoLevels = Assert<Equal<FieldArrayPathsAt<{a:{rows:{id:number}[]}; root:{id:number}[]}, 1>, "a.rows" | "root">>;
type Deep = {a:{b:{c:{d:{e:{f:{g:{h:{i:{j:{id:number}[]; k:{rows:{id:number}[]}}}}}}}}}}};
type Depth = Assert<Equal<FieldArrayPaths<Deep>, NativeArrayPaths<Deep>>>;
type ExplicitDepth = Assert<Equal<FieldArrayPathsAt<Deep, 4>, NativeArrayPathsAt<Deep, 4>>>;
type LastSegment = Assert<"a.b.c.d.e.f.g.h.i.j" extends FieldArrayPaths<Deep> ? true : false>;
type Truncated = Assert<"a.b.c.d.e.f.g.h.i.k.rows" extends FieldArrayPaths<Deep> ? false : true>;
type Deferred<T extends object> = FieldArrayPaths<{ payload:T[] }>;
type Specialized = Assert<Equal<Deferred<{id:number}>, "payload">>;
declare function register<T extends object>(value:T, path:FieldArrayPaths<T>): T;
const value = register({rows:[{id:1}]}, "rows");
type Inference = Assert<Equal<typeof value, {rows:{id:number}[]}>>;
function forward<T extends object>(value:T, path:FieldArrayPaths<T>) { return register<T>(value,path); }
