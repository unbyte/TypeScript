import { direct, nested, forwarded, specialized, stored, immediate, local } from "./out/library";
import type { Open } from "./out/provider";
declare module "./out/provider" { interface Open { second: number } }
type Assert<T extends true> = T;
type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type directResult = Assert<Equal<typeof direct, 2>>;
type nestedResult = Assert<Equal<typeof nested, 2>>;
type storedResult = Assert<Equal<typeof stored, 2>>;
type immediateResult = Assert<Equal<typeof immediate, 2>>;
const again = forwarded<Open>();
const againSpecialized = specialized();
type Forwarded = Assert<Equal<typeof again, 2>>;
type Specialized = Assert<Equal<typeof againSpecialized, 2>>;
type Local = Assert<Equal<typeof local, 2>>;
