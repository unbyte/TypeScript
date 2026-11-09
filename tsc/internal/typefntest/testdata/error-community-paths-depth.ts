import type { FieldPaths, FieldPathsAt } from "./paths";
type Bad = FieldPathsAt<{a:{b:string}}, 10>;
