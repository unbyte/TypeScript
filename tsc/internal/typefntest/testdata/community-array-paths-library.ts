import type { FieldArrayPaths } from "./paths";
export interface Form { main: {id:number}[] }
declare function paths<T extends object>(): FieldArrayPaths<T>;
export const result = paths<Form>();
export type Paths = FieldArrayPaths<Form>;
export type Deferred<T extends object> = FieldArrayPaths<T>;
