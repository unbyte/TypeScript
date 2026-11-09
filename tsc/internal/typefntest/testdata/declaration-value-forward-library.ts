import { read, annotated, inferred, type Open } from "./provider";
const source = read<Open>();
type alias = "unrelated type namespace";
const alias = source;
export const forwarded = alias;
export const fromAnnotation = annotated;
export const fromImport = inferred;
export const nested = { value: alias };
export class Box { readonly value = alias; get current() { return alias; } }
export function result() { return alias; }
const native = 1 as const;
export const nativeCopy = native;
export const widened: number = alias;
export function direct() { return read<Open>(); }
export const arrow = () => alias;
