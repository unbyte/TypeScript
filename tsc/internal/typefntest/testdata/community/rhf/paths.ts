// FieldPath traversal for form-object inputs. See ORIGIN.md and LICENSE.
import type { Primitive, BrowserNativeObject, OpaqueType, IsEqual } from "./contracts";

typefn SeenBefore<Seen, Value> {
  for (type previous in Seen) {
    if (type(IsEqual<previous, Value>) is true) { return true; }
  }
  return false;
}

typefn Edge<Key extends string | number, Value, Seen, Depth>: type {
  type result = never;
  for (type member in Value) {
    type key = `${Key}`;
    result = result | key;
    // A conditional with any takes both branches in the original algorithm.
    if (!reflect.isAny<member> && member is Primitive | BrowserNativeObject | OpaqueType) {
      continue;
    }
    if (SeenBefore<Seen, member>) { continue; }
    type children = Walk<member, Seen | member, Depth - 1>;
    result = result | `${Key}.${children}`;
  }
  return result;
}

typefn Walk<Value, Seen, Depth>: type {
  if (Depth < 0) { return never; }
  if (reflect.isAny<Value>) { return string; }
  if (!(Value is object)) { return never; }
  type keys = type(keyof Value);
  if (Value is type(ReadonlyArray<unknown>)) {
    type isArray = type(Value extends readonly unknown[] ? number extends Value["length"] ? true : false : false);
    if (isArray) { return Edge<number, Value[number], Seen, Depth>; }
    keys = type(Exclude<keyof Value, keyof unknown[]>);
  }
  type result = never;
  for (type key in keys) {
    if (key is string) { result = result | Edge<key, Value[key], Seen, Depth>; }
  }
  return result;
}

export typefn FieldPaths<Value extends object>: type {
  type result = never;
  for (type member in Value) { result = result | Walk<member, member, 9>; }
  return result;
}

// Deliberate algorithmic truncation, distinct from a VM resource failure.
export typefn FieldPathsAt<Value extends object, Depth extends number>: type {
  if (!numeric.isInteger<Depth> || Depth < 0 || Depth > 9) { fail<"Path depth must be an integer from 0 through 9">; }
  type result = never;
  for (type member in Value) { result = result | Walk<member, member, Depth>; }
  return result;
}
