import type { Remap, ComputedRemap } from "./library";
type Remapped = Remap<{ good: 2; bad: 1 }>;
type Computed = ComputedRemap<{ good: 2; bad: 1 }>;
type Valid = Remap<{ good: 2 }>;
typefn Properties<T> { return reflect.properties<T>; }
type First = Properties<Remapped /* first */>;
type Good = Properties<Valid>;
type Second = Properties<Remapped /* second */>;
type ComputedFirst = Properties<Computed /* first */>;
type ComputedSecond = Properties<Computed /* second */>;
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type ValidValue = Assert<Equal<Good[0]["value"], 2>>;
type ValidKey = Assert<Equal<Good[0]["key"], "good">>;
