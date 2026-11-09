import { result } from "./out/library";
declare module "./out/library" { interface Open { second: number } }
type Assert<T extends true> = T;
type Probe = Assert<typeof result extends string ? true : false>;
