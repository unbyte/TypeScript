// FieldPathValue and FieldArrayPathValue. See ORIGIN.md and LICENSE.
import type { FieldPaths } from "./paths";
import type { FieldArrayPaths } from "./array-paths";

// Keep native conditional narrowing attached to its generic binder. The native
// checker does not accept all equivalent checks over concrete capture aliases.
type Access<Value, Key> = Key extends keyof Value ? [true, Value[Key]] : [false, never];

typefn Resolve<Value, Path>: type {
  type result = never;
  for (type member in Value) {
    for (type path in Path) {
      // Native template inference is intentional: a symbolic `${string}.id`
      // must retain native matching behavior, not scan through an opaque hole.
      // The tagged alternatives also retain both native branches for any.
      type segments = type(path extends `${infer Head}.${infer Tail}` ? [true, Head, Tail] : [false, path, ""]);
      for (type segment in segments) {
        type head = segment[1];
        type tail = segment[2];
        type accesses = type(Access<member, head>);
        for (type access in accesses) {
          if (access[0]) {
            type child = access[1];
            if (segment[0]) {
              result = result | Resolve<child, tail>;
              if (undefined is child) { result = result | undefined; }
            } else {
              result = result | child;
            }
          } else {
            if (head is type(`${number}`)) {
              if (member is type(ReadonlyArray<unknown>)) {
                type child = member[number];
                if (segment[0]) { result = result | Resolve<child, tail>; }
                else { result = result | child; }
              } else {
                if (!segment[0] && member is undefined) { result = result | undefined; }
              }
            }
          }
        }
      }
    }
  }
  return result;
}

export typefn FieldPathValue<Value extends object, Path extends FieldPaths<Value>>: type {
  return Resolve<Value, Path>;
}

export typefn FieldArrayPathValue<Value extends object, Path extends FieldArrayPaths<Value>>: type {
  return Resolve<Value, Path>;
}
