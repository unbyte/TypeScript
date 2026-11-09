import type { Mapped, Input } from "./library";
type Target = { value: number; extra: 1 };
declare const bad: Mapped<Input> & { extra: 0 };
const first: Target = bad;
const second: Target = bad;
const unrelated: number = "wrong";
