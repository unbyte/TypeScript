# A complete parser as a design exercise

This is candidate R0 source, not executable code in this branch. It tests whether
the proposed syntax can express an ordinary recursive-descent parser without
type-level accumulator tricks. The result contract follows the inspected JSON
challenge: objects, arrays, strings, true, false and null; **numbers are rejected**.
Escaped quote, slash, backslash and b/f/n/r/t are supported. Unicode escape syntax
is deliberately outside this restricted parser, though literal UTF-16 text is
preserved. It is not a standard JSON implementation.

The native entry returns never on parse failure to match the challenge. Helpers
retain structured failures, so a different entry could report the failing offset
and expected token instead. Duplicate object keys use the last value. Trailing
commas and trailing input are rejected.

```ts
data Parsed = { value: Type; rest: string };
data ParseError = { rest: string; expected: string };
data ParseResult = Result<Parsed, ParseError>;

typefn space(text: string): string {
  return text.trimStart(" \t\r\n");
}

// Input begins immediately after the opening quote.
typefn jsonString(source: string): ParseResult {
  let rest = source;
  let value = "";
  while (rest != "") {
    match (rest) {
      `"${tail}` => {
        return ok({ value: literal(value), rest: tail });
      }
      `\\${escaped}${tail}` => {
        match (escaped) {
          '"' => { value += '"'; }
          "\\" => { value += "\\"; }
          "/" => { value += "/"; }
          "b" => { value += "\b"; }
          "f" => { value += "\f"; }
          "n" => { value += "\n"; }
          "r" => { value += "\r"; }
          "t" => { value += "\t"; }
          _ => { return err({ rest: rest, expected: "a supported escape" }); }
        }
        rest = tail;
      }
      `${ch}${tail}` => {
        if (ch == "\\") {
          return err({ rest: rest, expected: "an escape after backslash" });
        }
        if (text.codeUnitAt(ch, 0) < 32) {
          return err({ rest: rest, expected: "an escaped control character" });
        }
        value += ch;
        rest = tail;
      }
      _ => { return err({ rest: rest, expected: "a string character" }); }
    }
  }
  return err({ rest: rest, expected: "a closing quote" });
}

// Input begins immediately after the opening bracket.
typefn jsonArray(source: string): ParseResult {
  let rest = space(source);
  let items: List<Type> = [];
  match (rest) {
    `]${tail}` => { return ok({ value: type([]), rest: tail }); }
    _ => {}
  }
  while (true) {
    match (jsonValue(rest)) {
      ok(parsed) => { items.push(parsed.value); rest = space(parsed.rest); }
      err(error) => { return err(error); }
    }
    match (rest) {
      `]${tail}` => {
        return ok({
          value: tuple({ elements: items.map(x => element(x)), readonly: false }),
          rest: tail
        });
      }
      `,${tail}` => { rest = space(tail); }
      _ => { return err({ rest: rest, expected: "a comma or closing bracket" }); }
    }
  }
}

// Input begins immediately after the opening brace.
typefn jsonObject(source: string): ParseResult {
  let rest = space(source);
  let entries: Map<string, Type> = map();
  match (rest) {
    `}${tail}` => { return ok({ value: type({}), rest: tail }); }
    _ => {}
  }
  while (true) {
    let key = "";
    match (rest) {
      `"${tail}` => {
        match (jsonString(tail)) {
          ok(parsed) => { key = literal.text(parsed.value); rest = space(parsed.rest); }
          err(error) => { return err(error); }
        }
      }
      _ => { return err({ rest: rest, expected: "a quoted object key" }); }
    }
    match (rest) {
      `:${tail}` => { rest = space(tail); }
      _ => { return err({ rest: rest, expected: "a colon" }); }
    }
    match (jsonValue(rest)) {
      ok(parsed) => { entries.set(key, parsed.value); rest = space(parsed.rest); }
      err(error) => { return err(error); }
    }
    match (rest) {
      `}${tail}` => {
        let fields: List<Field> = [];
        for (const entry of entries.entries()) {
          fields.push(field(literal(entry.key), entry.value));
        }
        return ok({ value: object({ fields: fields }), rest: tail });
      }
      `,${tail}` => { rest = space(tail); }
      _ => { return err({ rest: rest, expected: "a comma or closing brace" }); }
    }
  }
}

typefn jsonValue(source: string): ParseResult {
  const input = space(source);
  match (input) {
    `"${rest}` => { return jsonString(rest); }
    `[${rest}` => { return jsonArray(rest); }
    `{${rest}` => { return jsonObject(rest); }
    `true${rest}` => { return ok({ value: type(true), rest: rest }); }
    `false${rest}` => { return ok({ value: type(false), rest: rest }); }
    `null${rest}` => { return ok({ value: type(null), rest: rest }); }
    _ => { return err({ rest: input, expected: "a supported JSON value" }); }
  }
}

typefn Parse<S> {
  const text = literal.text(S);
  match (jsonValue(text)) {
    ok(parsed) => {
      if (space(parsed.rest) != "") { return never; }
      return parsed.value;
    }
    err(_) => { return never; }
  }
}
```

## Expected cases

The nine inspected challenge cases require the following:

1. A nested object with string, boolean, null and mixed nested-array fields gives
   the corresponding native object/tuple Type.
2. `{}` gives `{}`.
3. `[]` gives `[]`.
4. `[1]` gives never under this challenge's deliberate number restriction.
5. `true` gives true.
6. `["Hello", true, false, null]` gives the corresponding four-element tuple.
7. An object key containing escaped r/n/b/f becomes a key containing those actual
   control units.
8. `{ 1: "world" }` gives never because object keys must be strings.
9. A raw newline inside a quoted key gives never.

Additional design oracles: `[true,]`, `{"x":false,}`, `truex`, a missing quote,
an incomplete escape and non-whitespace trailing input all produce never. A
single escaped backslash or quote is accepted. Duplicate object keys retain the
last value. Deep nesting consumes the shared call-depth account and produces a
resource diagnostic rather than never when the configured limit is exceeded.

The `true`/`false`/`null` prefix arms intentionally leave validation of the next
token to the enclosing delimiter/full-input check. This composes without accepting
`truex`. Every successful list/object iteration consumes input. Captures and parser
results are immutable snapshots; helper recursion never shares a mutable cursor.

## What this exercise changes in the design

It requires named data result kinds, helper recursion, ordinary strings/maps/lists,
explicit Type fields inside data, match captures with precise scope, constructors
for dynamic keys and tuple elements, and a distinction between parser failure and
VM failure. These are language requirements, not patches to one clever conditional
type. A standards JSON parser would additionally implement the number grammar and
Unicode escape/surrogate policy; a GraphQL or ArkType parser adds different grammar
and semantic phases rather than reusing this restricted parser by name.
