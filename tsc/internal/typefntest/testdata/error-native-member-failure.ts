import type { Mapped, Input, Read } from "./library";
type First = Read<Input /* first */>;
type Second = Read<Input /* second */>;
type DirectFirst = Mapped<Input /* direct first */>["value"];
type DirectSecond = Mapped<Input /* direct second */>["value"];
type Assert<T extends true> = T;
type SuccessfulSibling = Assert<Read<{value: 1}> extends 1 ? true : false>;
