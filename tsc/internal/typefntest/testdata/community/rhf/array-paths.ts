// FieldArrayPath traversal for form-object inputs. See ORIGIN.md and LICENSE.
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
    if (reflect.isAny<member>) { return string; }
    if (member is Primitive | BrowserNativeObject | OpaqueType) { continue; }
    if (SeenBefore<Seen, member>) { continue; }
    if (member is type(ReadonlyArray<unknown>)) {
      // The upstream element conditional distributes separately. Empty never[]
      // and primitive-only arrays contribute no path; any[] takes both branches.
      type hasObjectElement = false;
      for (type element in member[number]) {
        if (reflect.isAny<element> || !(element is Primitive | BrowserNativeObject | OpaqueType)) {
          hasObjectElement = true;
          break;
        }
      }
      if (!hasObjectElement) { continue; }
      result = result | `${Key}`;
    }
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

export typefn FieldArrayPaths<Value extends object>: type {
  type result = never;
  for (type member in Value) { result = result | Walk<member, member, 9>; }
  return result;
}

export typefn FieldArrayPathsAt<Value extends object, Depth extends number>: type {
  if (!numeric.isInteger<Depth> || Depth < 0 || Depth > 9) { fail<"Path depth must be an integer from 0 through 9">; }
  type result = never;
  for (type member in Value) { result = result | Walk<member, member, Depth>; }
  return result;
}
