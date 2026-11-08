# Syntax

Normative R2 chapter. The host is a TypeScript source/declaration file. The native
parser recognizes typefn declarations and delegates their bodies to the DSL parser.
The DSL parser never falls back to parsing a body as JavaScript.

## 1. Lexical rules

Whitespace and line/block comments separate tokens without semantic significance.
Identifiers use the host's identifier/Unicode escape rules; escaped reserved words
do not bypass keyword restrictions. Keywords are contextual outside a typefn.
Inside one, `type`, `typefn`, `return`, `if`, `else`, `while`, `for`, `of`, `in`,
`match`, `is`, `break`, `continue`, `extends`, `readonly`, `true`, `false`, `null`
and `undefined` have their specified syntactic roles. Native primitive Type names
remain built-in names. `_` is reserved for wildcard/constructor-slot notation.

String literals and template chunks use host escape decoding and preserve UTF-16
code units, including lone surrogates. Number/bigint tokens use host literal
spelling, including separators and radix prefixes. A sign is a unary operator.
Nonfinite number literals diagnose. Regex literals, tagged templates and JSX are
not DSL tokens. `/` means division or begins a comment.

Semicolons are mandatory after binding, assignment, return, application, break and
continue statements. There is no automatic semicolon insertion. Braces are
mandatory on every control-flow arm. Empty statements are not allowed. The scanner
treats `>` as individual closing tokens in DSL angle lists; `>=` is a comparison
token only outside that closing context. `>>`/`>>>` are not shift operators.

## 2. Core grammar

EBNF notation: `?`, `*`, `+` mean optional, repeated, and one-or-more; terminals
are quoted. `NativeType` is the host Type grammar, not its expression grammar.
`Name` is a nonreserved identifier; `MemberName` also admits keyword spellings.
Comma-separated lists allow one trailing comma. Empty slot lists denote arity zero.

```text
NamedFunction  = 'typefn' Name Parameters ReturnKind? Block ;
Lambda         = 'typefn' Parameters ReturnKind? Block ;
Parameters     = '<' (Parameter (',' Parameter)* ','?)? '>' ;
Parameter      = Name ('extends' TypeArgument)? | Name Slots ;
Slots          = '<' ('_' (',' '_')*)? '>' ;
ReturnKind     = ':' ('type' | 'typefn' Slots) ;

Block          = '{' Statement* '}' ;
Statement      = Binding | Assignment | Return | ApplyStatement
               | If | While | For | Match | NamedFunction | Block
               | 'break' ';' | 'continue' ';' ;
Binding        = 'type' Name '=' Expression ';' ;
Assignment     = Name '=' Expression ';' ;
Return         = 'return' Expression ';' ;
ApplyStatement = Application ';' ;
If             = 'if' '(' Expression ')' Block ('else' (Block | If))? ;
While          = 'while' '(' Expression ')' Block ;
For            = 'for' '(' 'type' Name ('of' | 'in') Expression ')' Block ;
Match          = 'match' '(' Expression ')' '{' Arm+ '}' ;
Arm            = Pattern ('if' '(' Expression ')')? '=>' Block ;

Expression     = precedence expression over Primary, Prefix and Binary ;
Primary        = Literal | QualifiedName | Application | Lambda | NativeIsland
               | '(' Expression ')' | Tuple | Object | Template ;
Postfix        = Primary ('[' Expression ']')* ;
Prefix         = ('!' | '+' | '-' | 'keyof') Prefix | Postfix ;
QualifiedName  = Name ('.' MemberName)* ;
Application    = QualifiedName '<' (TypeArgument (',' TypeArgument)* ','?)? '>' ;
NativeIsland   = 'type' '(' NativeType ')' ;
Tuple          = '[' (Element (',' Element)* ','?)? ']' ;
Element        = '...'? Expression ;
Object         = '{' (Field ((';' | ',') Field)* (';' | ',')?)? '}' ;
Field          = 'readonly'? FieldName '?'? ':' Expression ;
FieldName      = MemberName | StringLiteral | NumberLiteral | '[' Expression ']' ;
Template       = '`' (Chunk | '${' Expression '}')* '`' ;
Literal        = StringLiteral | NumberLiteral | BigintLiteral
               | 'true' | 'false' | 'null' | 'undefined' ;

Pattern        = '_' | Name | Literal | '-' (NumberLiteral | BigintLiteral)
               | TemplatePattern | TuplePattern | ObjectPattern
               | 'type' '(' NativeType ')' | 'is' TypeArgument ;
TuplePattern   = '[' (Pattern (',' Pattern)* (',' '...' NameOrWildcard)?
               | '...' NameOrWildcard)? ']' ;
ObjectPattern  = '{' (StaticKey ':' Pattern (',' StaticKey ':' Pattern)* ','?)? '}' ;
TemplatePattern= '`' (Chunk | '${' NameOrWildcard '}'
               | '${=' Expression '}')* '`' ;
NameOrWildcard = Name | '_' ;
StaticKey      = MemberName | StringLiteral | NumberLiteral ;
```

Binary operators and TypeArgument mode are specified below. Object/tuple expressions
are type construction, not JavaScript object/array expressions. In a field,
readonly followed by `:` or `?` is the property name; followed by another field
name it is a modifier. Keyword property names such as `type` and `this` are legal.
Computed field keys
must evaluate to one native property key. Tuple spreads must be tuple/array Types
and are checked using native variadic-tuple legality. Optional/labeled/rest tuple
syntax beyond these productions uses a native island.

Inside a body, bare names resolve only Types, constructors or intrinsic namespaces.
Dot qualification is symbol lookup, not object property access or host method
dispatch. Use `record["field"]` for a computed object's property. There are no
property assignments, destructuring, compound assignments, increment, ternaries,
ordinary parentheses calls, parameter defaults/rest, runtime declarations or casts.

Module-level `export typefn` and type-only import/export forms use native module
grammar. A native nongeneric type alias may additionally name a constructor.
Native conditional/mapped/function/constructor types, type queries, indexed/array
types and generic signatures may appear inside a native island. Within that island,
generic arguments use native Type syntax; scalar DSL expressions require a local
binding computed before entering the island. Native binders shadow captured names.
DSL Type/constructor locals and named typefns are available through the lexical
bridge. Prelude intrinsic namespaces are available only in DSL expressions;
evaluate an intrinsic into a local before referring to its result in an island.

```ts
typefn Callback<T> { return type((value: T) => Promise<T>); }
typefn Optional<T> { return type({ [K in keyof T]?: T[K] }); }
```

## 3. Precedence and evaluation order

From tightest to loosest:

| Level | Operators | Associativity |
| --- | --- | --- |
| 9 | Name application, indexing | Indexing chains left; application callee is a name |
| 8 | `! + - keyof` | Right |
| 7 | `* / %` | Left |
| 6 | `+ -` | Left |
| 5 | `&` | Left, native Type intersection |
| 4 | `\|` | Left, native Type union |
| 3 | `is == != < <= > >=` | Nonassociative |
| 2 | `&&` | Left, short-circuit |
| 1 | `\|\|` | Left, short-circuit |

Operands, application arguments, tuple elements and object fields evaluate in
source order. `&&` and `||` evaluate their right operand only when necessary.
`A is B | C` means `A is (B | C)`. Comparison chains diagnose, including mixed
operators; explicit grouping allows comparing the result of a comparison.

## 4. Angle application versus comparison

This is a syntactic ordered-choice rule, independent of declarations and whitespace:

1. After a qualified name followed by `<`, attempt Application recognition.
2. At the outer level of each TypeArgument, parse only precedence levels 4–9.
   Thus arithmetic, unions/intersections and nested applications are admitted;
   comparisons and logical operators require parentheses. Parenthesized groups,
   tuple elements, object field values and index operands use full Expression.
3. Recognition succeeds only if the entire list closes and the next token is a
   valid application follower: EOF, `;`, `,`, `)`, `]`, `}`, `[`, a binary operator,
   or an enclosing angle-list close. In the TypeArgument of an `is` pattern only,
   `if` and `=>` also terminate that pattern and are valid followers. A bare
   identifier, literal, `(`, `{` or `:` is not a follower. Template interpolation
   `}` is an ordinary close. The extra pattern-follow mode is syntactic and finite.
4. If recognition succeeds, commit to Application. Otherwise restore the token
   position and parse the name as an ordinary expression followed by comparison.
   Definition/kind errors never trigger syntactic backtracking.
5. Within an angle list, an unparenthesized `>` closes the list. `>=` in that
   position is not a predicate argument: parenthesize the comparison. The scanner
   must preserve enough token offsets to recover after speculative close splitting.

```ts
F<A + 1, G<B>>;             // Application.
type p = A < B && C > D;    // Two comparisons.
type q = Choose<(A < B)>;   // Predicate argument.
type r = F < T >;           // Same call as F<T>.
// A < B > C                // Rejected comparison chain.
// F<T><U>                  // No chained constructor application.
```

The grammar therefore chooses a parse rather than promising that every possible
relational spelling is convenient. Bind an intermediate constructor before applying
it. Outside a body, ordinary TypeScript generic argument syntax is unchanged.

An implementation MUST memoize speculative recognition by token offset, grammar
mode and precedence (or provide equivalent bounded work); it MUST NOT rescan nested
candidate suffixes exponentially. Recognition cannot consult the binder, checker
or VM. Native islands use the host Type parser with a bounded island entry/exit,
not repeated host parsing attempts. Nesting and recognition-work limits apply.

## 5. Pattern and recovery rules

Exactly one final, unguarded `_` arm is required. It is the only wildcard arm;
other whole-value capture arms are permitted and may make later arms unreachable.
Names in a pattern are fresh bindings; existing names are compared with `type(Name)`
or `is Name`. Repeated capture names in one pattern diagnose. Tuple rest capture is
last. Template `${=expression}` pins a string literal computed in the outer scope;
it never refers to a capture from that same arm.

The parser switches explicitly between DSL, template pattern and native Type
island modes. Delimiters inside decoded string/template chunks do not terminate
their containing construct. Invalid code produces diagnostics/recovery nodes;
recovery nodes MUST NOT be executed. Recovery synchronizes at a statement semicolon
or balanced block/arm boundary, retaining source spans for the editor.
