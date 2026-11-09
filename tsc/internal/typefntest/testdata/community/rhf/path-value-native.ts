// Original conditional path-value lookup. See ORIGIN.md and LICENSE.
import type { Path } from "./paths-native";
import type { ArrayPath } from "./array-paths-native";
type ArrayKey = number;
type PathValueImpl<T, P extends string> = T extends unknown
  ? P extends `${infer K}.${infer R}`
    ? K extends keyof T
      ? undefined extends T[K]
        ? PathValueImpl<T[K], R> | undefined
        : PathValueImpl<T[K], R>
      : K extends `${ArrayKey}`
        ? T extends ReadonlyArray<infer V>
          ? PathValueImpl<V, R>
          : never
        : never
    : P extends keyof T
      ? T[P]
      : P extends `${ArrayKey}`
        ? T extends ReadonlyArray<infer V>
          ? V
          : T extends undefined
            ? undefined
            : never
        : never
  : never;

export type NativePathValue<T extends object, P extends Path<T> | ArrayPath<T>> = PathValueImpl<T, P>;
