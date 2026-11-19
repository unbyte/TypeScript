//// [tests/cases/compiler/typeFunctionTemplatePatterns.ts] ////

//// [typeFunctionTemplatePatterns.ts]
export typefn Greeting<T extends string> {
    match T {
        case `hello ${infer C}`: { return C; }
        else: { return undefined; }
    }
}

export declare const name: Greeting<"hello Ada">;
export declare const empty: Greeting<"hello ">;
export declare const absent: Greeting<"goodbye">;
const ada: "Ada" = name;
const nothing: "" = empty;
const missing: undefined = absent;


//// [typeFunctionTemplatePatterns.js]
const ada = name;
const nothing = empty;
const missing = absent;
export {};


//// [typeFunctionTemplatePatterns.d.ts]
export typefn Greeting<T extends string> {
    match T {
        case `hello ${infer C}`: {
            return C;
        }
        else: {
            return undefined;
        }
    }
}
export declare const name: Greeting<"hello Ada">;
export declare const empty: Greeting<"hello ">;
export declare const absent: Greeting<"goodbye">;
