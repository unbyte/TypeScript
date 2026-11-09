import type { FieldPaths, FieldPathsAt } from "./paths";
declare function register<T extends object>(value:T, path:FieldPaths<T>): T;
register({user:{id:1}}, "user.missing");
