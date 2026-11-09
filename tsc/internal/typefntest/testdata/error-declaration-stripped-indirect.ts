/** @internal */
type Hidden = { hidden: true };
type Indirect = Hidden;
export typefn Public<T> { return [T, Indirect]; }
