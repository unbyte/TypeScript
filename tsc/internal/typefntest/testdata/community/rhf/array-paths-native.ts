// Original conditional array-path traversal. See ORIGIN.md and LICENSE.
import type { Primitive, BrowserNativeObject, OpaqueType, IsEqual } from "./contracts";
type ArrayKey = number;
type IsAny<T> = 0 extends 1 & T ? true : false;
type IsTuple<T extends ReadonlyArray<unknown>> = number extends T["length"] ? false : true;
type TupleKeys<T extends ReadonlyArray<unknown>> = Exclude<keyof T, keyof unknown[]>;
type Prev = [never, 0, 1, 2, 3, 4, 5, 6, 7, 8, ...0[]];
type AnyIsEqual<T1, T2> = T1 extends T2
  ? IsEqual<T1, T2> extends true ? true : never
  : never;

type ArrayPathImpl<
  K extends string | number,
  V,
  TraversedTypes,
  D extends number = 9,
> = V extends Primitive | BrowserNativeObject | OpaqueType
  ? IsAny<V> extends true
    ? string
    : never
  : V extends ReadonlyArray<infer U>
    ? U extends Primitive | BrowserNativeObject | OpaqueType
      ? IsAny<V> extends true
        ? string
        : never
      : // Check so that we don't recurse into the same type
        // by ensuring that the types are mutually assignable
        // mutually required to avoid false positives of subtypes
        true extends AnyIsEqual<TraversedTypes, V>
        ? never
        : `${K}` | `${K}.${ArrayPathInternal<V, TraversedTypes | V, Prev[D]>}`
    : true extends AnyIsEqual<TraversedTypes, V>
      ? never
      : `${K}.${ArrayPathInternal<V, TraversedTypes | V, Prev[D]>}`;

/**
 * Helper type for recursively constructing paths through a type.
 * This obscures the internal type param TraversedTypes from exported contract.
 *
 * See {@link ArrayPath}
 */
type ArrayPathInternal<T, TraversedTypes = T, D extends number = 9> = [
  D,
] extends [never]
  ? never
  : T extends ReadonlyArray<infer V>
    ? IsTuple<T> extends true
      ? {
          [K in TupleKeys<T>]-?: ArrayPathImpl<
            K & string,
            T[K],
            TraversedTypes,
            D
          >;
        }[TupleKeys<T>]
      : ArrayPathImpl<ArrayKey, V, TraversedTypes, D>
    : {
        [K in keyof T]-?: ArrayPathImpl<K & string, T[K], TraversedTypes, D>;
      }[keyof T];

/**
 * Type which eagerly collects all paths through a type which point to an array
 * type.
 * @typeParam T - type which should be introspected.
 * @example
 * ```
 * Path<{foo: {bar: string[], baz: number[]}}> = 'foo.bar' | 'foo.baz'
 * ```
 */
// We want to explode the union type and process each individually
// so assignable types don't leak onto the stack from the base.
export type ArrayPath<T> = T extends unknown ? ArrayPathInternal<T> : never;

export type NativeArrayPaths<T extends object> = ArrayPath<T>;
export type NativeArrayPathsAt<T extends object, D extends number> = T extends unknown ? ArrayPathInternal<T, T, D> : never;
