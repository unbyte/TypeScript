import type { Mapped, Input } from "./out/library";
type Target = { value: number };
declare const bad: Mapped<Input>;
const first: Target = bad /* first */;
const second: Target = bad /* second */;
const good: Target = { value: 1 };
declare const valid: Mapped<{value: 1}>;
const successfulSibling: Target = valid;
const unrelated: number = "wrong";
