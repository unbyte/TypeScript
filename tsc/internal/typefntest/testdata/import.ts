import type { Double } from "./calculations";
import type { Double as Reexported } from "./reexport";
const a: Double<21> = 42;
const b: Reexported<3> = 6;
export { a, b };
