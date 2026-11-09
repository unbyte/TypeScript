import type { WordsLiteral } from "./words";
type Bad = WordsLiteral<"fooBar" & {readonly brand: unique symbol}, true, false>;
