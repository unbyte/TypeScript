import type { Read, ReadKeys, ReadComputed } from "./library";
type Bad = { bad: 1; good: 2 };
type First = Read<Bad /* first */>;
type Second = Read<Bad /* second */>;
type ConstraintFirst = ReadKeys<"bad" /* first */>;
type ConstraintSecond = ReadKeys<"bad" /* second */>;
type ComputedFirst = ReadComputed<Bad /* first */>;
type ComputedSecond = ReadComputed<Bad /* second */>;
