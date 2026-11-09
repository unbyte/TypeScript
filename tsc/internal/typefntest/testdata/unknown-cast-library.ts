typefn Increment<N> { return N + 1; }
typefn Checked<N> { type used = Increment<N>; return string; }
type Unknown = unknown;
export function fromUnknown<N>(value:unknown):Increment<N> { return value as Increment<N>; }
export function toUnknown<N>(value:Increment<N>):unknown { return value as unknown; }
export function fromAlias<N>(value:Unknown):Increment<N> { return <Increment<N>>value; }
export function toAlias<N>(value:Increment<N>):Unknown { return <Unknown>value; }
export function guarded<N>(value:unknown):Checked<N> { return value as Checked<N>; }
export function eraseGuard<N>(value:Checked<N>):unknown { return value as unknown; }
export function doubleCast<N>(value:string):Increment<N> { return value as unknown as Increment<N>; }
