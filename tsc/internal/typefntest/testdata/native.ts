import type { Box } from "./helper.js";
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends
  (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type Result = Assert<Equal<Box<string>, { value: string }>>;
export const answer: 42 = 42;
