# Worked design workloads

These are complete candidate-language examples, **not executed tests**. They
exercise R0's notation and contracts before implementation. Expected results are
design oracles. Library names indicate the source of requirements, not a claim
to have ported that library. Source provenance is in the research manifest.

## 1. Query-string parser: ordered deduplication

This implements the contract of the 15 inspected type-challenges query-string
tests. Bare keys become true, repeated distinct values become ordered tuples,
identical repeats are discarded. The added policy is to ignore empty segments.
There is no URL decoding or special handling of `+`.

```ts
typefn ParseQueryString<S> {
  const text = literal.text(S);
  let values: Map<string, Type> = map();

  for (const part of text.split("&")) {
    if (part == "") { continue; }
    let key = part;
    type value = true;
    match (part) {
      `${name}=${rest}` => {
        key = name;
        value = literal(rest);
      }
      _ => {}
    }

    match (values.get(key)) {
      none => { values.set(key, value); }
      some(previous) => {
        if (reflect.kind(previous) == "tuple") {
          const shape = reflect.tuple(previous);
          let found = false;
          for (const item of shape.elements) {
            if (same(item.type, value)) { found = true; break; }
          }
          if (!found) { values.set(key, type([...previous, value])); }
        } else {
          if (!same(previous, value)) {
            values.set(key, type([previous, value]));
          }
        }
      }
    }
  }

  let fields: List<Field> = [];
  for (const entry of values.entries()) {
    fields.push(field(literal(entry.key), entry.value));
  }
  return object({ fields: fields });
}
```

| Input | Expected result |
| --- | --- |
| `''` | `{}` |
| `'k1'` | `{k1:true}` |
| `'k1&k1'` | `{k1:true}` |
| `'k1&k2'` | `{k1:true,k2:true}` |
| `'k1=v1'` | `{k1:'v1'}` |
| `'k1=v1&k1=v2'` | `{k1:['v1','v2']}` |
| `'k1=v1&k2=v2'` | `{k1:'v1',k2:'v2'}` |
| `'k1=v1&k2=v2&k1=v2'` | `{k1:['v1','v2'],k2:'v2'}` |
| `'k1=v1&k2'` | `{k1:'v1',k2:true}` |
| `'k1=v1&k1=v1'` | `{k1:'v1'}` |
| `'k1=v1&k1=v2&k1=v1'` | `{k1:['v1','v2']}` |
| `'k1=v1&k2=v1&k1=v2&k1=v1'` | `{k1:['v1','v2'],k2:'v1'}` |
| `'k1=v1&k2=v2&k1=v2&k1=v3'` | `{k1:['v1','v2','v3'],k2:'v2'}` |
| `'k1=v1&k1'` | `{k1:['v1',true]}` |
| `'k1&k1=v1'` | `{k1:[true,'v1']}` |

Additional oracles: `a=b=c` stores `b=c`; `=x` uses an empty key; `a=` stores an
empty string; `&&a&&` produces `{a:true}`. Broad string is a literal-extraction
diagnostic. An unresolved S suspends; a union requires explicit member mapping.
The parser deliberately stores values as Types but its cursor, keys and map are
ordinary data. There is no tuple-length arithmetic or inference accumulator.

## 2. String matching without a fixed delimiter

```ts
typefn Split<S, Separator> {
  const text = literal.text(S);
  const separator = literal.text(Separator);
  if (separator == "") { fail("Split requires a nonempty separator"); }
  let rest = text;
  let parts: List<Type> = [];
  while (true) {
    match (rest) {
      `${head}${=separator}${tail}` => {
        parts.push(literal(head));
        rest = tail;
      }
      _ => {
        parts.push(literal(rest));
        return tuple({ elements: parts.map(x => element(x)), readonly: false });
      }
    }
  }
}
```

`Split<'a::b::','::'>` is `['a','b','']`. The nonempty check proves each successful
iteration consumes input. Pinned delimiters do not capture or resolve a type
variable by accident. A regex is unnecessary for this operation.

This does not by itself implement type-fest Words: acronym/case/number policies,
brand handling and broad-string fallbacks remain separate algorithms. Those
policies should be expressed as explicit data branches, not hidden in match.

## 3. Descriptor round trips and symbolic Types

```ts
typefn ReadonlyShape<T> {
  let shape = reflect.object(T);
  for (let i = 0; i < shape.fields.length; i += 1) {
    shape.fields[i].readonly = true;
  }
  for (let i = 0; i < shape.indexers.length; i += 1) {
    shape.indexers[i].readonly = true;
  }
  return object(shape);
}

typefn Box<T> { return { value: T }; }

function read<T>(box: Box<T>): T { return box.value; }
declare function inferBox<T>(box: Box<T>): T;
const inferred = inferBox({ value: 123 }); // Native inference decides widening.
```

`ReadonlyShape<{a?: string}>` must retain optional absence under both settings of
exactOptionalPropertyTypes. Calls/constructs/indexers are not accidentally dropped.
An origin-restricted class cannot be cloned and edited silently. The generic
function body must work before T is concrete. Testing only `Box<string>` misses
the central integration requirement.

## 4. Overload-preserving async conversion

```ts
typefn asyncSignature(signature: Signature): Signature {
  match (signature) {
    { kind: "generic" } => {
      fail("AsyncMethod requires concrete signatures; preserve generic methods explicitly");
    }
    { kind: "concrete" } => {
      match (signature.predicate) {
        some(_) => { fail("An async result cannot preserve this type predicate"); }
        none => {}
      }
      return { ...signature, returns: type(Promise<signature.returns>) };
    }
  }
}

typefn AsyncMethod<T> {
  const signatures = reflect.calls(T);
  if (signatures.length == 0) { fail("Expected a callable type"); }
  return overload(signatures.map(asyncSignature));
}
```

All overloads, their order, parameters and this are retained. Predicates cannot
be copied unchanged into a Promise-returning function. Existing generic
signatures are reported rather than accidentally reduced to their last overload
or losing relationships. An origin-restricted signature may still reject editing;
the constructor is the final validity boundary.

This example exposes why complete reflection needs more than Parameters/ReturnType.
It is not a full ts-toolbelt Curry port; generative polymorphic currying is outside
the agreed construction scope.

## 5. Correlated unions and explicit products

```ts
typefn PairProducts<A, B> {
  type result = never;
  for (const left of members(A)) {
    for (const right of members(B)) {
      result = result | [left, right];
    }
  }
  return result;
}

typefn IndexEvents<Events> {
  type result = never;
  for (const event of members(Events)) {
    type Tagged = { [K in event["type"]]: event };
    result = result | Tagged;
  }
  return result;
}
```

`PairProducts<'a'|'b',1|2>` has four tuple members. Work/output quotas apply to
larger products. `IndexEvents<{type:'a';x:number}|{type:'b';y:string}>` yields a
union of separate correlated records, not one record losing discriminator/value
relationships. Combining those records into a lookup object requires an explicit
intersection or key policy. This illustrates part of ts-pattern/XState's needs;
it does not claim that a product alone implements deep exclusion or negation.

## 6. Higher-order composition

```ts
typefn pipe(input: Type, stages: List<(value: Type) => Type>): Type {
  type result = input;
  for (const stage of stages) { result = stage(result); }
  return result;
}

typefn Box<T> { return { value: T }; }
typefn Freeze<T> { return { readonly [K in keyof T]: T[K] }; }

typefn FrozenBox<T> { return pipe(T, [Box, Freeze]); }
```

The callable values carry kind signatures and immutable captures. Arity/kind
errors are static. `FrozenBox<string>` is `{readonly value:string}`. This replaces
hotscript's internal this/rawArgs encoding without generating a new native generic
declaration. Open fp-ts registries can still be referenced through native type
expressions; reflection does not freeze module augmentation into a closed clone.

## 7. Schema and SQL directionality

```ts
typefn ColumnRoles<Select, Insert, Update> {
  return { select: Select; insert: Insert; update: Update };
}

typefn SchemaTypes<Input, Output, Requirements> {
  return { input: Input; output: Output; requirements: Requirements };
}

typefn LeftJoined<T> { return T | null; }
```

These minimal constructors intentionally retain independent directions. A parser
may accept a string and output a number; an effectful schema also carries context.
`LeftJoined<{id:number;name:string}>` makes the whole row nullable, which is not
`{id:number|null;name:string|null}`. A real SQL selection algorithm must carry
alias and nullability metadata until it knows which boundary becomes nullable.

The validation corpus must embed these results in generic callbacks, fluent
builders and NoInfer-sensitive argument positions. There is no credible inference
compatibility claim from these three constructors alone.

## 8. Recursive shapes and enum identity

```ts
typefn Tree<Value> {
  return recursive((self: Type) => type({
    value: Value;
    children: readonly self[];
  }));
}

typefn Status<> {
  return enumType("status", [
    { name: "Open", value: "open" },
    { name: "Closed", value: "closed" }
  ]);
}
```

`Tree<T>` preserves a real graph, not a recursion depth truncation. `Status` denotes
a native type-only enum, without a runtime `Status.Open` value. Repeated evaluation
uses the same specified identity. The descriptor's native member Types can build
additional object shapes. Structural literal compatibility should instead use
enumValues. Declaration emit must preserve identity and guarded recursion before
either capability is advertised.

## 9. Parser architecture

[The complete restricted JSON parser](json-parser.md) exercises ordinary data
state, recursive helpers, string captures, errors, nested arrays/objects and final
input consumption. Its scope intentionally matches the inspected challenge's
number rejection. It is a language-design exercise, not a standards JSON claim.

ArkType/gql.tada require richer lexing, precedence, directives, scopes, fragments,
schema lookup, completion/error information and different semantic phases. R0 has
the data and Type operations to write those algorithms without tuple arithmetic;
actual ports and upstream fixtures are still required to demonstrate coverage.
