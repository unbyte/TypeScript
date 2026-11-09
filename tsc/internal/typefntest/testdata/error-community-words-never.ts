import type { WordsLiteral } from "./words";
type Bad = WordsLiteral<never, true, false>;
