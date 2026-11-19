// @strict: true
// @target: esnext

type Digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9";

typefn Skip<S> {
    type Rest = S;
    while (true) {
        match Rest {
            case ` ${infer Tail}`: { Rest = Tail; }
            else: { return Rest; }
        }
    }
}

typefn Atom<S> {
    type Rest = Skip<S>;
    match Rest {
        case `-${infer Tail}`: {
            type [Value, After] = Atom<Tail>;
            return [number.negate<Value>, After];
        }
        case `(${infer Tail}`: {
            type [Value, After] = Sum<Tail>;
            match Skip<After> {
                case `)${infer End}`: { return [Value, End]; }
                else: { fail<"expected closing parenthesis">; }
            }
        }
        else: {} // Scan the integer at the start of Rest.
    }
    type Digits = "";
    while (true) {
        match Rest {
            case `${infer D extends Digit}${infer Tail}`: {
                Digits = `${Digits}${D}`;
                Rest = Tail;
            }
            else: { break; }
        }
    }
    if (Digits === "") { fail<"expected an integer or parenthesis">; }
    return [number.parse<Digits>, Rest];
}

typefn Product<S> {
    type [Value, Rest] = Atom<S>;
    while (true) {
        match Skip<Rest> {
            case `*${infer Right}`: {
                type [Term, Next] = Atom<Right>;
                Value = number.multiply<Value, Term>;
                Rest = Next;
            }
            else: { return [Value, Rest]; }
        }
    }
}

typefn Sum<S> {
    type [Value, Rest] = Product<S>;
    while (true) {
        match Skip<Rest> {
            case `${infer Op extends "+" | "-"}${infer Right}`: {
                type [Term, Next] = Product<Right>;
                if (Op === "+") { Value = number.add<Value, Term>; }
                else { Value = number.subtract<Value, Term>; }
                Rest = Next;
            }
            else: { return [Value, Rest]; }
        }
    }
}

typefn Calc<S extends string> {
    type [Value, Rest] = Sum<S>;
    if (Skip<Rest> !== "") { fail<"trailing input">; }
    return Value;
}
declare const precedence: Calc<"2 + 3 * (8 - 5)">;
declare const associativity: Calc<"10 - 3 - 2">;
declare const unary: Calc<"-(2 + 3) * 4">;
const eleven: 11 = precedence;
const five: 5 = associativity;
const negative: -20 = unary;
// @ts-expect-error precedence yields the literal 11
const wrong: 42 = precedence;
declare const multiDigit: Calc<" 12 + 34 ">;
const fortySix: 46 = multiDigit;
