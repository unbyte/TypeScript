import type { FieldPaths, FieldPathsAt } from "./paths";
type Bad = FieldPaths<{a:{b:{c:string}}}>;
