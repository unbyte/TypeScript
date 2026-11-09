import type { Alias, ThroughAlias, Nested } from "./library";
type First = ThroughAlias<0 /* first */>;
type Second = ThroughAlias<0 /* second */>;
type DirectFirst = Alias<0 /* direct first */>;
type DirectSecond = Alias<0 /* direct second */>;
type NestedFirst = Nested<0 /* nested first */>;
type NestedSecond = Nested<0 /* nested second */>;
type Assert<T extends true> = T;
type SuccessfulSibling = Assert<ThroughAlias<1> extends 1 ? true : false>;
