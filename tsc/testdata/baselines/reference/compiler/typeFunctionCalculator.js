//// [tests/cases/compiler/typeFunctionCalculator.ts] ////

//// [typeFunctionCalculator.ts]
type Cursor<S extends string, I extends number = 0> = { source: S; offset: I };

typefn Skip<C> {
    match string.match<C, { run: [[" ", " "]] }> {
        case ["hit", infer Text, infer Next] { return Next; }
        else { return C; }
    }
}

typefn Atom<C> {
    type At = Skip<C>;
    match string.match<At, { run: [["0", "9"]] }> {
        case ["hit", infer Digits, infer Next] { return [number.fromDecimal<Digits>, Next]; }
    }
    match string.read<At> {
        case ["hit", "-", infer Next] {
            type [Value, After] = Atom<Next>;
            return [number.negate<Value>, After];
        }
        case ["hit", "(", infer Next] {
            type [Value, After] = Sum<Next>;
            match string.match<Skip<After>, ")"> {
                case ["hit", ")", infer End] { return [Value, End]; }
            }
            fail<"expected closing parenthesis">;
        }
    }
    fail<"expected an integer or parenthesis">;
}

typefn Product<C> {
    type [Value, At] = Atom<C>;
    while (true) {
        match string.match<Skip<At>, "*"> {
            case ["hit", "*", infer Right] {
                type [Term, Next] = Atom<Right>;
                Value = number.multiply<Value, Term>;
                At = Next;
            }
            else { break; }
        }
    }
    return [Value, At];
}

typefn Sum<C> {
    type [Value, At] = Product<C>;
    while (true) {
        match string.read<Skip<At>> {
            case ["hit", infer Op, infer Right] {
                if (Op !== "+" && Op !== "-") { break; }
                type [Term, Next] = Product<Right>;
                if (Op === "+") { Value = number.add<Value, Term>; }
                else { Value = number.subtract<Value, Term>; }
                At = Next;
            }
            else { break; }
        }
    }
    return [Value, At];
}

typefn Calc<S extends string> {
    type [Value, At] = Sum<Cursor<S>>;
    match string.read<Skip<At>> {
        case ["eof"] { return Value; }
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
