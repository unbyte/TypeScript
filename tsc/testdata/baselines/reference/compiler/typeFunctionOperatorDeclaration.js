//// [tests/cases/compiler/typeFunctionOperatorDeclaration.ts] ////

//// [typeFunctionOperatorDeclaration.ts]
typefn Make<N> { return [<T>(arg: T) => T + N]; }
declare const functions: Make<3>;
export const named: Make<3>[0] = functions[0];
export const inferred = functions[0];
export const computed = named(4 as const);


//// [typeFunctionOperatorDeclaration.js]
export const named = functions[0];
export const inferred = functions[0];
export const computed = named(4);
