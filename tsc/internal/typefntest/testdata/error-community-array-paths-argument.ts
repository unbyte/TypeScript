import type { FieldArrayPaths, FieldArrayPathsAt } from "./paths";
declare function register<T extends object>(value:T, path:FieldArrayPaths<T>): T;
register({rows:[{id:1}], strings:["a"]}, "strings");
