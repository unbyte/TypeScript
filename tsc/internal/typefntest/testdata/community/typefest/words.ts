// Executable port of type-fest's finite-literal word scanner. This entry point
// takes resolved options; public Words brand normalization is a separate policy.
// See ORIGIN.md, words-native.ts and LICENSE in this directory.
import type { WordSeparators, AsciiPunctuation } from "./characters";

export typefn WordsLiteral<Sentence extends string, Numbers extends boolean, Punctuation extends boolean> {
  if (reflect.kind<Sentence> != "stringLiteral") {
    fail<"WordsLiteral requires one unbranded string literal">;
  }
  type result = [];
  type start = 0;
  type cursor = 0;
  type last = "";

  for (type character of Sentence) {
    if (character is WordSeparators || (Punctuation is true && character is AsciiPunctuation)) {
      if (start < cursor) { result = [...result, text.slice<Sentence, start, cursor>]; }
      cursor = cursor + 1;
      start = cursor;
      last = "";
      continue;
    }
    if (last == "") {
      last = character;
      cursor = cursor + 1;
      continue;
    }

    type wasNumeric = text.digit<last> >= 0n;
    type isNumeric = text.digit<character> >= 0n;
    type split = false;
    if (wasNumeric != isNumeric) {
      split = Numbers is true;
    } else {
      if (!isNumeric) {
        type wasLower = last is type(Lowercase<last>);
        type wasUpper = last is type(Uppercase<last>);
        type isLower = character is type(Lowercase<character>);
        type isUpper = character is type(Uppercase<character>);
        if (wasLower && isUpper && !isLower) {
          split = true;
        } else {
          if (wasUpper && isLower && !wasLower && !isUpper) {
            type boundary = cursor - 1;
            if (start < boundary) { result = [...result, text.slice<Sentence, start, boundary>]; }
            start = boundary;
            last = character;
            cursor = cursor + 1;
            continue;
          }
        }
      }
    }
    if (split) {
      if (start < cursor) { result = [...result, text.slice<Sentence, start, cursor>]; }
      start = cursor;
    }
    last = character;
    cursor = cursor + 1;
  }
  if (start < cursor) { result = [...result, text.slice<Sentence, start, cursor>]; }
  return result;
}

// Distribution is explicit and preserves each member's independent scanner state.
export typefn WordsUnion<Sentence extends string, Numbers extends boolean, Punctuation extends boolean> {
  type result = never;
  for (type member in Sentence) {
    result = result | WordsLiteral<member, Numbers, Punctuation>;
  }
  return result;
}
