import type { FieldArrayPaths, FieldArrayPathsAt } from "./paths";
type Bad = FieldArrayPaths<{a:{b:{c:string}}}>;
