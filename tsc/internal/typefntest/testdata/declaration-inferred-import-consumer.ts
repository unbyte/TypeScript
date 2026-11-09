import {count} from "./out/library";
declare module "./out/provider" { interface Open { second: number } }
type Assert<T extends true> = T;
type Result = Assert<typeof count extends 2 ? true : false>;
