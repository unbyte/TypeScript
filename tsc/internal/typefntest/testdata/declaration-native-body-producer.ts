/** @internal */
type Hidden = 1;
function runtime(): 1 {
 const unused: Hidden = 1;
 return 1;
}
export typefn Signature<> { return type(typeof runtime); }
const value: number = 1 as Hidden;
export typefn Value<> { return type(typeof value); }
