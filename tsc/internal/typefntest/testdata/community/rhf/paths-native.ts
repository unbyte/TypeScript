// Original conditional traversal, using the selected upstream contracts.
import type { Primitive, BrowserNativeObject, OpaqueType, IsEqual } from "./contracts";
type ArrayKey = number;
type IsTuple<T extends ReadonlyArray<unknown>> = number extends T["length"] ? false : true;
type TupleKeys<T extends ReadonlyArray<unknown>> = Exclude<keyof T, keyof unknown[]>;
type Prev = [never, 0, 1, 2, 3, 4, 5, 6, 7, 8, ...0[]];
type AnyIsEqual<T1, T2> = T1 extends T2
  ? IsEqual<T1, T2> extends true
    ? true
    : never
  : never;

/**
 * Helper type for recursively constructing paths through a type.
 * This actually constructs the strings and recurses into nested
 * object types.
 *
 * See {@link Path}
 */
type PathImpl<
  K extends string | number,
  V,
  TraversedTypes,
  D extends number = 9,
> = V extends Primitive | BrowserNativeObject | OpaqueType
  ? `${K}`
  : // Check so that we don't recurse into the same type
    // by ensuring that the types are mutually assignable
    // mutually required to avoid false positives of subtypes
    true extends AnyIsEqual<TraversedTypes, V>
    ? `${K}`
    : `${K}` | `${K}.${PathInternal<V, TraversedTypes | V, Prev[D]>}`;

/**
 * Helper type for recursively constructing paths through a type.
 * This obscures the internal type param TraversedTypes from exported contract.
 *
 * See {@link Path}
 */
type PathInternal<T, TraversedTypes = T, D extends number = 9> = [D] extends [
  never,
]
  ? never
  : T extends ReadonlyArray<infer V>
    ? IsTuple<T> extends true
      ? {
          [K in TupleKeys<T>]-?: PathImpl<K & string, T[K], TraversedTypes, D>;
        }[TupleKeys<T>]
      : PathImpl<ArrayKey, V, TraversedTypes, D>
    : {
        [K in keyof T]-?: PathImpl<K & string, T[K], TraversedTypes, D>;
      }[keyof T];

/**
 * Type which eagerly collects all paths through a type
 * @typeParam T - type which should be introspected
 * @example
 * ```
 * Path<{foo: {bar: string}}> = 'foo' | 'foo.bar'
 * ```
 */
// We want to explode the union type and process each individually
// so assignable types don't leak onto the stack from the base.
export type Path<T> = T extends unknown ? PathInternal<T> : never;


export type NativePaths<T extends object> = Path<T>;
export type NativePathsAt<T extends object, D extends number> = T extends unknown ? PathInternal<T, T, D> : never;
