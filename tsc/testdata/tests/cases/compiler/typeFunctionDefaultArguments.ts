// @strict: true
// @target: esnext

typefn Pair<T, U = T> { return [T, U]; }
typefn Chain<T = "first", U = [T], V = [T, U]> { return V; }
typefn Forward<X> { return Pair<X>; }

declare const same: Pair<1>;
declare const different: Pair<1, 2>;
declare const allDefaults: Chain<>;
declare const someDefaults: Chain<2, "second">;
const a: [1, 1] = same;
const b: [1, 2] = different;
const c: ["first", ["first"]] = allDefaults;
const d: [2, "second"] = someDefaults;

type Later<T> = Forward<T>;
declare const x: Later<"x">;
declare const y: Later<"y">;
const xPair: ["x", "x"] = x;
const yPair: ["y", "y"] = y;
type Swapped<A, B extends string = "label"> = Pair<B, A>;
declare const swapped: Swapped<7>;
const labeled: ["label", 7] = swapped;

type Box<T> = { value: T };
typefn Wrap<F<_>, T extends string | number, U = F<T>> { return U; }
declare const box: Wrap<Box<_>, 3>;
const three: 3 = box.value;

type Entry<K, V> = { key: K; value: V };
typefn WithBox<T = 1, F<_> = Box<_>, G<_> = F<_>> { return G<T>; }
typefn WithKey<K, F<_> = Entry<K, _>> { return F<"value">; }
declare const defaultBox: WithBox<>;
declare const overrideBox: WithBox<2, Entry<"custom", _>>;
declare const defaultKey: WithKey<"key">;
const one: 1 = defaultBox.value;
const custom: { key: "custom"; value: 2 } = overrideBox;
const keyed: { key: "key"; value: "value" } = defaultKey;

typefn Derived<T extends string, U = Wrap<Box<_>, T>> { return U; }
declare const derived: Derived<"nested">;
const nested: "nested" = derived.value;
type NativeDefault<T = Wrap<Box<_>, "native">> = T;
declare const native: NativeDefault;
const literal: "native" = native.value;

typefn Unused<> { fail<"overridden default was evaluated">; }
typefn Select<T = Unused<>> { return T; }
declare const override: Select<4>;
const four: 4 = override;

typefn Checked<T extends string, U extends T = T> { return U; }
declare const checked: Checked<"ok">;
const ok: "ok" = checked;
// @ts-expect-error explicit overrides must respect the constraint
type BadConstraint = Checked<"ok", "other">;
// @ts-expect-error a required argument is missing
type TooFew = Pair<>;
// @ts-expect-error excess arguments are rejected
type TooMany = Pair<1, 2, 3>;
// @ts-expect-error applying all defaults still requires <>
type Bare = Chain;
// @ts-expect-error constructor defaults must have the declared number of holes
typefn WrongConstructor<F<_> = Entry<_, _>> { return F<0>; }
// @ts-expect-error defaults can only refer to earlier parameters
typefn LaterConstructor<F<_> = G<_>, G<_> = Box<_>> { return F<0>; }
// @ts-expect-error a required parameter cannot follow a default
typefn Misordered<T = 1, U> { return [T, U]; }
// @ts-expect-error the default must satisfy its own constraint
typefn InvalidDefault<T extends number = "text"> { return T; }
typefn InvalidBody<> {
    // @ts-expect-error arity is checked even in an unused body
    return Pair<>;
}
