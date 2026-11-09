import type { FieldArrayPaths, FieldArrayPathsAt } from "./paths";
type Bad = FieldArrayPathsAt<{a:{b:string}}, 10>;
