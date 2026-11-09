import type { FieldPathValue, FieldArrayPathValue } from "./values";
type Missing = FieldPathValue<{a:1}, "missing">;
type PrimitiveArray = FieldArrayPathValue<{strings:string[]}, "strings">;
type MixedPath = FieldPathValue<{a:1}, "a"|"missing">;
type InvalidInput = FieldPathValue<string, any>;
