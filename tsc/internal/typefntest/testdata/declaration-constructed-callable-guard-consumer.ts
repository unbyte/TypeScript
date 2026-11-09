import { result } from "./out/library";
declare module "./out/library" { interface Open { second: number } }
export const probe: typeof result = 1;
