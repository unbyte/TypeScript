import type { FieldPaths, FieldPathsAt } from "./paths";
import type { NativePaths, NativePathsAt } from "./reference";
type Assert<T extends true> = T;
type Equal<A,B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
interface Node { value: string; next?: Node }
interface Left { right: Right; tag: string }
interface Right { left: Left; count: number }
declare const symbol: unique symbol;
class Nominal { private brand!: void; nested!: {name: string}; }
type Cases = [
 {}, {foo:{bar:string}}, {optional?:{value:number}}, {value:never}, {value?:never},
 {value:unknown}, {value:any}, {date:Date; files:FileList; file:File}, Node, Left,
 readonly [first:{name:string}, second?:{count:number}], {items:{name:string}[]},
 readonly [head:{a:string}, ...tail:{b:number}[]], {0:{a:string}; "1":{b:string}; [symbol]:{secret:string}},
 Record<string,{value:number}>, {[key:number]:{value:number}}, Nominal,
 {a:{x:string}} | {b:{y:number}}, {a:{x:string}} | {a:{x:string}; b:number},
 {"a.b":{c:string}}, {fn: (() => void) & {meta:{name:string}}}, never[], [], any, never
];
type Compare0 = Assert<Equal<FieldPaths<Cases[0]>, NativePaths<Cases[0]>>>;
type Compare1 = Assert<Equal<FieldPaths<Cases[1]>, NativePaths<Cases[1]>>>;
type Compare2 = Assert<Equal<FieldPaths<Cases[2]>, NativePaths<Cases[2]>>>;
type Compare3 = Assert<Equal<FieldPaths<Cases[3]>, NativePaths<Cases[3]>>>;
type Compare4 = Assert<Equal<FieldPaths<Cases[4]>, NativePaths<Cases[4]>>>;
type Compare5 = Assert<Equal<FieldPaths<Cases[5]>, NativePaths<Cases[5]>>>;
type Compare6 = Assert<Equal<FieldPaths<Cases[6]>, NativePaths<Cases[6]>>>;
type Compare7 = Assert<Equal<FieldPaths<Cases[7]>, NativePaths<Cases[7]>>>;
type Compare8 = Assert<Equal<FieldPaths<Cases[8]>, NativePaths<Cases[8]>>>;
type Compare9 = Assert<Equal<FieldPaths<Cases[9]>, NativePaths<Cases[9]>>>;
type Compare10 = Assert<Equal<FieldPaths<Cases[10]>, NativePaths<Cases[10]>>>;
type Compare11 = Assert<Equal<FieldPaths<Cases[11]>, NativePaths<Cases[11]>>>;
type Compare12 = Assert<Equal<FieldPaths<Cases[12]>, NativePaths<Cases[12]>>>;
type Compare13 = Assert<Equal<FieldPaths<Cases[13]>, NativePaths<Cases[13]>>>;
type Compare14 = Assert<Equal<FieldPaths<Cases[14]>, NativePaths<Cases[14]>>>;
type Compare15 = Assert<Equal<FieldPaths<Cases[15]>, NativePaths<Cases[15]>>>;
type Compare16 = Assert<Equal<FieldPaths<Cases[16]>, NativePaths<Cases[16]>>>;
type Compare17 = Assert<Equal<FieldPaths<Cases[17]>, NativePaths<Cases[17]>>>;
type Compare18 = Assert<Equal<FieldPaths<Cases[18]>, NativePaths<Cases[18]>>>;
type Compare19 = Assert<Equal<FieldPaths<Cases[19]>, NativePaths<Cases[19]>>>;
type Compare20 = Assert<Equal<FieldPaths<Cases[20]>, NativePaths<Cases[20]>>>;
type Compare21 = Assert<Equal<FieldPaths<Cases[21]>, NativePaths<Cases[21]>>>;
type Compare22 = Assert<Equal<FieldPaths<Cases[22]>, NativePaths<Cases[22]>>>;
type Compare23 = Assert<Equal<FieldPaths<Cases[23]>, NativePaths<Cases[23]>>>;
type Compare24 = Assert<Equal<FieldPaths<Cases[24]>, NativePaths<Cases[24]>>>;
type Fixed = Assert<Equal<FieldPaths<{user:{name:string}; items:{id:number}[]}>, "user" | "user.name" | "items" | `items.${number}` | `items.${number}.id`>>;
type Recursive = Assert<Equal<FieldPaths<Node>, "value" | "next">>;
type Shallow = Assert<Equal<FieldPathsAt<{a:{b:{c:string}}}, 0>, "a">>;
type TwoLevels = Assert<Equal<FieldPathsAt<{a:{b:{c:string}}}, 1>, "a" | "a.b">>;
type Deferred<T extends object> = FieldPaths<{payload:T}>;
type Specialized = Assert<Equal<Deferred<{id:number}>, "payload" | "payload.id">>;
type Deep = {a:{b:{c:{d:{e:{f:{g:{h:{i:{j:{k:{l:string}}}}}}}}}}}};
type DefaultDepth = Assert<Equal<FieldPaths<Deep>, NativePaths<Deep>>>;
type ExplicitDepth = Assert<Equal<FieldPathsAt<Deep, 4>, NativePathsAt<Deep, 4>>>;

type LastSegment = Assert<"a.b.c.d.e.f.g.h.i.j" extends FieldPaths<Deep> ? true : false>;
type Truncated = Assert<"a.b.c.d.e.f.g.h.i.j.k" extends FieldPaths<Deep> ? false : true>;
declare function register<T extends object>(value: T, path: FieldPaths<T>): T;
const value = register({user:{id:1}}, "user.id");
type Inference = Assert<Equal<typeof value, {user:{id:number}}>>;
function forward<T extends object>(value:T, path:FieldPaths<T>) {
  return register<T>(value, path);
}
