import type { FieldPathValue } from "./values";
type Exhausted = FieldPathValue<any, "a.b.c.d.e">;
