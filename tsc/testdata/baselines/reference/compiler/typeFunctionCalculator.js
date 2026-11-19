//// [tests/cases/compiler/typeFunctionCalculator.ts] ////

//// [typeFunctionCalculator.ts]
typefn Skip<S, Index> {
    match string.match<S, Index, { run: [[" ", " "]] }> {
        case undefined { return Index; }
        case [infer Text, infer Next] { return Next; }
    }
}

typefn Atom<S, Index> {
    type At = Skip<S, Index>;
    match string.match<S, At, { run: [["0", "9"]] }> {
        case undefined {} // Try a unary operator or a parenthesized expression.
        case [infer Digits, infer Next] { return [number.fromDecimal<Digits>, Next]; }
    }
    match string.read<S, At> {
        case undefined { fail<"expected an integer or parenthesis">; }
        case ["-", infer Next] {
            type [Value, After] = Atom<S, Next>;
            return [number.negate<Value>, After];
        }
        case ["(", infer Next] {
            type [Value, After] = Sum<S, Next>;
            match string.match<S, Skip<S, After>, ")"> {
                case undefined { fail<"expected closing parenthesis">; }
                case [")", infer End] { return [Value, End]; }
            }
        }
        else { fail<"expected an integer or parenthesis">; }
    }
}

typefn Product<S, Index> {
    type [Value, At] = Atom<S, Index>;
    while (true) {
        match string.match<S, Skip<S, At>, "*"> {
            case undefined { break; }
            case ["*", infer Right] {
                type [Term, Next] = Atom<S, Right>;
                Value = number.multiply<Value, Term>;
                At = Next;
            }
        }
    }
    return [Value, At];
}

typefn Sum<S, Index> {
    type [Value, At] = Product<S, Index>;
    while (true) {
        match string.read<S, Skip<S, At>> {
            case undefined { break; }
            case [infer Op, infer Right] {
                if (Op !== "+" && Op !== "-") { break; }
                type [Term, Next] = Product<S, Right>;
                if (Op === "+") { Value = number.add<Value, Term>; }
                else { Value = number.subtract<Value, Term>; }
                At = Next;
            }
        }
    }
    return [Value, At];
}

typefn Calc<S extends string> {
    type [Value, At] = Sum<S, 0>;
    match string.read<S, Skip<S, At>> {
        case undefined { return Value; }
        else { fail<"trailing input">; }
    }
}
declare const precedence: Calc<"2 + 3 * (8 - 5)">;
declare const associativity: Calc<"10 - 3 - 2">;
declare const unary: Calc<"-(2 + 3) * 4">;
const eleven: 11 = precedence;
const five: 5 = associativity;
const negative: -20 = unary;
// @ts-expect-error precedence yields the literal 11
const wrong: 42 = precedence;


//// [typeFunctionCalculator.js]
"use strict";
const eleven = precedence;
const five = associativity;
const negative = unary;
// @ts-expect-error precedence yields the literal 11
const wrong = precedence;
