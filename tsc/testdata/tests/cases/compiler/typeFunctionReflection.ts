// @strict: true
// @exactOptionalPropertyTypes: true
// @target: esnext

declare const token: unique symbol;
class Model {
    private hidden = 0;
    readonly [token] = "tag" as const;
    get count(): number { return this.hidden; }
    set count(value: string | number) { this.hidden = Number(value); }
    run?(id: number): void;
}
typefn Fields<T> { return object.fields<T>; }
typefn Project<T> { return object.withFields<{}, object.fields<T>>; }
declare const fields: Fields<Model>;
const getter: [number] = fields.count.getter;
const setter: [number | string] = fields.count.setter;
const method: "method" = fields.run.kind;
const keyKind: "data" = fields[token].kind;
// @ts-expect-error private members do not appear in the projection
fields.hidden;
declare let projection: Project<Model>;
projection.count = "2";
const read: number = projection.count;
const tag: "tag" = projection[token];
// @ts-expect-error a symbol-keyed readonly field retains its modifier
projection[token] = "tag";

typefn TupleCopy<T> { return tuple.compose<tuple.decompose<T>>; }
declare const tuple: TupleCopy<readonly [head: number, tail?: string]>;
const head: number = tuple[0];

typefn Template<T> { return template.compose<{ texts: ["on", ""]; types: [T] }>; }
declare const eventName: Template<"Ready" | "Close">;
const events: "onReady" | "onClose" = eventName;

typefn Sort<T> { return union.toTuple<T>; }
declare const order: Sort<"😀" | "a" | "𐀀">;
const keys: ["a", "𐀀", "😀"] = order;
