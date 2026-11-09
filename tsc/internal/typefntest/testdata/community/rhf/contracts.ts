// Selected upstream leaf/registry contracts. See ORIGIN.md and LICENSE.
interface File extends Blob {
  readonly lastModified: number;
  readonly name: string;
}

interface FileList {
  readonly length: number;
  item(index: number): File | null;
  [index: number]: File;
}

export type Primitive =
  | null
  | undefined
  | string
  | number
  | boolean
  | symbol
  | bigint;

export type BrowserNativeObject = Date | FileList | File;

/**
 * Registry of types which the recursive type helpers ({@link Path},
 * {@link DeepPartial}, FieldErrors, ...) treat as opaque leaf values
 * instead of recursing into their properties.
 *
 * Empty by default, so it has no effect until a consumer registers a type
 * via declaration merging. Useful for rich third-party value types (Dayjs,
 * Decimal, Luxon DateTime, ...) whose members should never be addressed by
 * a form path and can be expensive for the compiler to traverse.
 * @example
 * ```
 * declare module 'react-hook-form' {
 *   interface OpaqueTypes {
 *     dayjs: Dayjs;
 *   }
 * }
 * ```
 */
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface OpaqueTypes {}

/**
 * Union of all types registered in {@link OpaqueTypes}. `never` while the
 * registry is empty.
 */
export type OpaqueType = OpaqueTypes[keyof OpaqueTypes];


export type IsEqual<T1, T2> = T1 extends T2
  ? (<G>() => G extends T1 ? 1 : 2) extends <G>() => G extends T2 ? 1 : 2
    ? true
    : false
  : false;
