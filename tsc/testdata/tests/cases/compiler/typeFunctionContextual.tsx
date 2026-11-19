// @strict: true
// @target: esnext
// @jsx: preserve

namespace JSX { export interface IntrinsicElements { div: {} } }
const typefn = <T,>(x: T) => x;
const ordinary: 1 = typefn(1);
const jsx = <div>{typefn("value")}</div>;
namespace Native {
    export type typefn<T> = { value: T };
    export typefn Box<T> { return { wrapped: typefn<T> }; }
}
declare const item: Native.Box<42>;
const result: 42 = item.wrapped.value;
