//// [tests/cases/compiler/typeFunctionCalculator.ts] ////

//// [typeFunctionCalculator.ts]
typefn Calc<S extends string> {
    type Tokens = Tokenize<S>;
    type [Value, End] = Sum<Tokens, 0>;
    if (Tokens[End] !== undefined) {
        fail<"unexpected token after expression">;
    }
    return Value;
}

type Digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9";
type Whitespace = " " | "\t" | "\n" | "\r";

typefn Tokenize<S> {
    type Tokens = [];
    type At = 0;
    type End = string.length<S>;
    while (At !== End) {
        type Start = At;
        type [Char, Next] = string.read<S, At>;
        At = Next;
        if (Char is Whitespace) {
            continue;
        }
        if (Char is Digit) {
            while (string.read<S, At> is [Digit, number]) {
                At = At + 1;
            }
            type Text = string.slice<S, Start, At>;
            Tokens = [...Tokens, number.parse<Text>];
        } else {
            Tokens = [...Tokens, Char];
        }
    }
    return [...Tokens, undefined];
}

// Each parser returns the value and the index of the next token.
typefn Sum<Tokens, Index> {
    type [Value, At] = Product<Tokens, Index>;
    while (Tokens[At] is "+" | "-") {
        type Op = Tokens[At];
        type [Right, Next] = Product<Tokens, At + 1>;
        if (Op === "+") {
            Value = Value + Right;
        } else {
            Value = Value - Right;
        }
        At = Next;
    }
    return [Value, At];
}

typefn Product<Tokens, Index> {
    type [Value, At] = Atom<Tokens, Index>;
    while (Tokens[At] === "*") {
        type [Right, Next] = Atom<Tokens, At + 1>;
        Value = Value * Right;
        At = Next;
    }
    return [Value, At];
}

typefn Atom<Tokens, Index> {
    type At = Index;
    type Negative = false;
    while (Tokens[At] === "-") {
        Negative = !Negative;
        At = At + 1;
    }
    type Value = Tokens[At];
    At = At + 1;
    match Value {
        case number: {}
        case "(": {
            type [Inner, Close] = Sum<Tokens, At>;
            if (Tokens[Close] !== ")") {
                fail<"expected closing parenthesis">;
            }
            Value = Inner;
            At = Close + 1;
        }
        else: { fail<"expected an integer or parenthesis">; }
    }
    if (Negative) {
        Value = -Value;
    }
    return [Value, At];
}

declare const precedence: Calc<"2 + 3 * (8 - 5)">;
declare const associativity: Calc<"10 - 3 - 2">;
declare const unary: Calc<"-(2 + 3) * 4">;
const eleven: 11 = precedence;
const five: 5 = associativity;
const negative: -20 = unary;
// @ts-expect-error precedence yields the literal 11
const wrong: 42 = precedence;
declare const multiDigit: Calc<" \t12 + 34\r\n">;
const fortySix: 46 = multiDigit;

// @ts-expect-error whitespace separates integer tokens
declare const separatedDigits: Calc<"1 2">;
const invalid = separatedDigits;


//// [typeFunctionCalculator.js]
"use strict";
const eleven = precedence;
const five = associativity;
const negative = unary;
// @ts-expect-error precedence yields the literal 11
const wrong = precedence;
const fortySix = multiDigit;
const invalid = separatedDigits;
