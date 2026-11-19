// @strict: true
// @target: esnext
// @declaration: true

typefn Make<N> { return [<T>(arg: T) => [T, N + 1]]; }
declare const functions: Make<3>;
export const inferred = functions[0];
export const computed = inferred("x" as const);

typefn Increment<N> { return N + 1; }
type Hidden<N> = { value: Increment<N> };
typefn Build<> { return [Hidden<3>]; }
declare const wrapped: Build<>;
export const object = wrapped[0];

typefn Transform<N> {
    return [<T>(arg: T) => T extends string ? { text: T; tag: N + 1 }
                                                : { readonly [K in keyof T]: [T[K], N + 1] }];
}
declare const transforms: Transform<3>;
export const transform = transforms[0];
