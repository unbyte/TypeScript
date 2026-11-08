> Historical R1 material, superseded by the [R2 specification](../../spec/README.md). Not the current language contract.

# String literal types: HTTP bindings

The earlier guide only showed a small `/users/${id}` match. This worked design
example covers a complete binding such as `GET /users/:id<number>/posts/:slug`.
The [complete source](examples/http-routes.typefn) contains every definition below.
It uses the proposed R1 syntax; the current compiler cannot execute it yet.

## Extract the method and route

This is the small, reusable string-Type operation:

```ts
typefn SplitBinding<S> {
  match (S) {
    `${method} ${route}` => {
      return { method: method; route: route };
    }
    _ => { return never; }
  }
}

type Parts = SplitBinding<"GET /users/:id">;
// { method: "GET"; route: "/users/:id" }

type Method = Parts["method"]; // "GET"
type Route = Parts["route"];   // "/users/:id"
```

Because S is a Type, the captures are **string literal Types**, not data strings.
There is no explicit infer declaration and no conditional-type recursion. The
space is a literal delimiter; matching splits at its first occurrence and the
final capture gets the rest. The entire subject is matched, without backtracking.

Extraction also has a useful symbolic-pattern case:

```ts
type Open = SplitBinding<`GET /users/${string}`>;
// { method: "GET"; route: `/users/${string}` }
```

The fixed prefix determines the first space and method; the route remains a
template Type. This does not enumerate possible URLs. An unresolved parameter S
waits if its split is unknown; it does not fall through to never. Closed broad
string has no guaranteed space-delimited structure, so SplitBinding<string> is
never under this pattern. Process unions explicitly.

SplitBinding is only structural extraction. It does not validate HTTP methods,
path syntax or parameter declarations. That belongs to the full parser below.

## Build the parameter object and handler Type

The complete example defines HttpBinding, RouteParams, RouteField and Param:

```ts
type Endpoint = HttpBinding<"GET /users/:id<number>/posts/:slug">;
// {
//   method: "GET";
//   route: "/users/:id<number>/posts/:slug";
//   params: { id: number; slug: string };
// }

type Method = Endpoint["method"];
type Route = Endpoint["route"];
type Params = Endpoint["params"];

type GetHandler = Handler<"GET /users/:id<number>">;
// (params: { id: number }) => Promise<void>
```

The main composition is small; the full source includes the validation helpers:

```ts
typefn HttpBinding<S> {
  const text = literal.text(S);
  if (text.length == 0) { fail("Expected METHOD /path"); }

  type parts = SplitBinding<S>;
  if (parts == never) { fail("Expected METHOD /path"); }
  type method = parts["method"];
  type route = parts["route"];

  if (method is HttpMethod) {
    return { method: method; route: route; params: RouteParams<route> };
  }
  fail("Unknown HTTP method; expected an uppercase method");
}
```

RouteParams scans segments using the pattern `${head}/${tail}` in a while loop.
RouteField matches typed parameters before untyped ones:

```ts
match (Segment) {
  `:${name}<${kind}>` => {
    match (kind) {
      "string" => { return Param<name, string>; }
      "number" => { return Param<name, number>; }
      "boolean" => { return Param<name, boolean>; }
      _ => { fail("Unknown route parameter type"); }
    }
  }
  `:${name}` => { return Param<name, string>; }
  `*${_}` => { fail("Wildcard route segments are not supported here"); }
  _ => { return {}; }
}
```

Param validates the name using ordinary compile-time string operations, then
constructs `{[K in Name]: Value}`. RouteParams checks duplicate keys, accumulates
the fields and returns a mapped object. Handler constructs an ordinary native
function Type from the resulting params. These are dynamically built Types;
no interface declaration or runtime function needs to be generated.

Typed parameters describe a router contract: this example assumes a runtime
router performs the corresponding number/boolean conversion. The DSL does not
convert URL text, install handlers or generate that router. Plain `:name` parameters
stay string. The `route` field preserves the original binding path, including its
annotations; it is not a newly generated runtime path pattern.

## Deliberately small route grammar

* Method is uppercase GET, POST, PUT, PATCH, DELETE, HEAD or OPTIONS.
* Exactly one ASCII space separates the method from an absolute path.
* `/` is valid. Other paths have nonempty slash-separated segments and no trailing slash.
* `:name`, `:name<string>`, `:name<number>` and `:name<boolean>` are parameter segments.
* Names are ASCII identifiers: initial letter, `_` or `$`; later digits are allowed.
* Names cannot repeat. Optional parameters and leading-`*` wildcard segments are excluded.
* Spaces, tabs, CR/LF, queries and fragments are rejected in the path. Percent escapes
  are retained as literal text; no URL decoding or regular-expression interpretation occurs.

This is a binding-language example, not a claim to implement HTTP or every routing
framework's grammar. Static segments are literal text; only a leading colon or
star invokes the special segment rules above.

HttpBinding requires a single unbranded string literal via literal.text. Therefore
it rejects broad string and infinite template Types, although SplitBinding can
extract known structure from some template Types. Unresolved generic input waits
and resumes under the [evaluation rules](evaluation.md).

## Preserve correlation across a union

```ts
typefn HttpBindings<S> {
  type result = never;
  for (type binding in S) {
    result = result | HttpBinding<binding>;
  }
  return result;
}

type Endpoints = HttpBindings<
  "GET /users/:id<number>" | "POST /users"
>;
// | { method: "GET"; route: "/users/:id<number>"; params: { id: number } }
// | { method: "POST"; route: "/users"; params: {} }

type Methods = Endpoints["method"]; // "GET" | "POST"
type GetRoutes = Extract<Endpoints, { method: "GET" }>["route"];
// "/users/:id<number>"
```

This keeps each method paired with its own route and parameters. Constructing
one object with separate method and route unions would admit combinations absent
from the input. Union order is irrelevant here. Duplicate input members normalize
normally; this does not impose uniqueness of route registrations across a router.

An invalid member makes HttpBindings fail. The code deliberately does not silently
filter invalid bindings to never. HttpBindings<never> is the empty union; passing
never directly to HttpBinding fails its single-literal requirement.

## Computed delimiters use the same matching operation

```ts
typefn SplitOnce<S, Delimiter> {
  const separator = literal.text(Delimiter);
  if (separator.length == 0) { fail("Expected a nonempty delimiter"); }
  match (S) {
    `${left}${=separator}${right}` => { return [left, right]; }
    _ => { return [S]; }
  }
}

type Pieces = SplitOnce<"a::b::c", "::">;
// ["a", "b::c"]
```

`${=separator}` pins a computed delimiter. `${left}` and `${right}` introduce
captures. A delimiter such as `.` or `*` is literal text, not a regex operator.

## Required outcomes

These are design cases, not passing parser/compiler tests. Use the full source's
definitions and separate invalid inputs into independent fixtures.

| ID | Expression/input | Expected |
| --- | --- | --- |
| H01 | SplitBinding<"GET /users/:id"> | Literal method GET and route /users/:id |
| H02 | SplitBinding<`GET /users/${string}`> | GET and the preserved template route |
| H03 | SplitBinding<"GET/users"> | never; extraction has no separator |
| H04 | SplitBinding<string> | never; broad string does not guarantee this pattern |
| H05 | HttpBinding<"GET /"> | GET, `/`, params {} |
| H06 | HttpBinding<"HEAD /health"> | HEAD, `/health`, params {} |
| H07 | HttpBinding<"GET /users/:id"> | params {id:string} |
| H08 | HttpBinding<"GET /users/:id<number>/posts/:slug"> | params {id:number;slug:string} |
| H09 | HttpBinding<"PATCH /flags/:enabled<boolean>"> | params {enabled:boolean} |
| H10 | HttpBinding<"GET /:id/:id"> | Duplicate-name diagnostic |
| H11 | HttpBinding<"get /users"> | Unsupported-method diagnostic; no implicit uppercase conversion |
| H12 | HttpBinding<"GET  /users"> | Reject extra separator space |
| H13 | HttpBinding<"GET users"> | Reject nonabsolute route |
| H14 | HttpBinding<"GET /users/"> or "GET /users//posts" | Reject trailing/empty segment |
| H15 | HttpBinding<"GET /users?q=x"> or "GET /users#top" | Reject query/fragment |
| H16 | HttpBinding<"GET /:id<uuid>"> | Unsupported parameter-type diagnostic |
| H17 | HttpBinding<"GET /:"> or "GET /:9id" | Empty/invalid identifier diagnostic |
| H18 | HttpBinding<"GET /:id?"> or "GET /*rest" | Reject optional/wildcard grammar |
| H19 | HttpBinding<string> or HttpBinding<`GET /${string}`> | Single-literal requirement fails; not an unresolved generic binder |
| H20 | HttpBinding<S> with unresolved S, even S extends string | Pending; substitute a valid literal to obtain its endpoint Type |
| H21 | HttpBindings of the two routes above | Correlated endpoint union; GetRoutes excludes the POST route |
| H22 | HttpBindings containing an invalid binding | Failure retained, not silently removed by union construction |
| H23 | HttpBindings<never> / HttpBinding<never> | never / single-literal diagnostic |
| H24 | Handler<"GET /users/:id<number>"> | Native function Type accepting {id:number} and returning Promise<void> |
| H25 | SplitOnce<"a::b::c","::"> | ["a","b::c"], first delimiter, no backtracking |
| H26 | SplitOnce<"abc","::"> / SplitOnce<"abc",""> | ["abc"] / empty-delimiter diagnostic |

For declaration emit, export Handler and a concrete constructor-independent alias,
retain their transitive helper definitions, and repeat these observations with
producer sources absent. Any failure must identify the binding and original
validation operation rather than an anonymous mapped-type error.
