import type { FieldPathValue } from "./values";
import type { FieldPaths } from "./paths";
export interface Form { user: {id:number} }
export type Values = FieldPathValue<Form, FieldPaths<Form>>;
declare function all<T extends object>(): FieldPathValue<T, FieldPaths<T>>;
export const result = all<Form>();
export type Deferred<T extends object, P extends FieldPaths<T>> = FieldPathValue<T,P>;
