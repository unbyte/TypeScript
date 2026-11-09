import type { FieldPaths } from "./paths";
export interface Form { user: { id:number } }
declare function paths<T extends object>(): FieldPaths<T>;
export const result = paths<Form>();
export type Paths = FieldPaths<Form>;
export type Deferred<T extends object> = FieldPaths<T>;
