import type { WordsLiteral } from "./words";
type Bad = WordsLiteral<"foo" | "bar", true, false>;
