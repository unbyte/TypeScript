export {};
type Assert<T extends true> = T;
type Equal<A,B> = (<T>()=>T extends A?1:2) extends (<T>()=>T extends B?1:2)?true:false;
typefn Snapshot<T> {
 type saved = T;
 typefn Pair<U> { return [saved,U]; }
 saved = "changed";
 return Pair<123>;
}
type SnapshotValue = Assert<Equal<Snapshot<"original">,["original",123]>>;
type SnapshotPending<T> = Snapshot<T>;
type SnapshotReplay = Assert<Equal<SnapshotPending<456>,[456,123]>>;
typefn Select<T> {
 type saved = T;
 typefn Choose<U> { if(saved == "yes") {return U;} return false; }
 saved = "no";
 return {value:Choose<123>};
}
type Pending<T> = Select<T>;
type Selected = Assert<Equal<Pending<"yes">,{value:123}>>;
type Rejected = Assert<Equal<Pending<"no">,{value:false}>>;
typefn Nested<T> {
 type prefix = T;
 typefn First<U> { return text.concat<[prefix,U]>; }
 typefn Second<U> { return First<U>; }
 prefix = "changed";
 return Second<"!">;
}
type NestedPending<T> = Nested<T>;
type NestedReplay = Assert<Equal<NestedPending<"hello">,"hello!">>;
typefn Deeper<T> {
 typefn First<U> {
  type local = U;
  typefn Second<V> { return [T,local,V]; }
  local = "changed";
  return Second<1>;
 }
 return First<2>;
}
type DeeperReplay = Assert<Equal<Deeper<3>,[3,2,1]>>;
typefn Sum<T> {
 typefn Add<N>: type { if(N == 0) {return T;} return N + Add<N - 1>; }
 return Add<5>;
}
type Recursive = Assert<Equal<Sum<10>,25>>;
type RecursivePending<T> = Sum<T>;
type RecursiveReplay = Assert<Equal<RecursivePending<100>,115>>;
typefn Iterations<T> {
 type output = [];
 for(type item of T) {
  typefn Copy<> {return item;}
  output = [...output,Copy<>];
 }
 return output;
}
type Iterated = Assert<Equal<Iterations<[1,2,3]>,[1,2,3]>>;
typefn Shadow<T> {
 typefn Helper<U> { type T = U; return T; }
 return Helper<123>;
}
type Shadowed = Assert<Equal<Shadow<"outer">,123>>;
typefn Native<T> {
 typefn Helper<U> { return type({captured:T; own:U}); }
 return Helper<123>;
}
type NativePending<T> = Native<T>;
type NativeReplay = Assert<Equal<NativePending<"x">,{captured:"x";own:123}>>;
typefn Constrained<T> {
 type bound = T;
 typefn Helper<U extends bound> {return U;}
 return Helper<"x">;
}
type Constraint = Assert<Equal<Constrained<string>,"x">>;
type ConstraintPending<T> = Constrained<T>;
type ConstraintReplay = Assert<Equal<ConstraintPending<string>,"x">>;
typefn ParameterShadow<T> {
 typefn Helper<Helper> {return Helper;}
 return Helper<123>;
}
type ParameterShadowsSelf = Assert<Equal<ParameterShadow<0>,123>>;

typefn DeferredExpression<T> {
 typefn Identity<U> {return U;}
 return text.concat<[Identity<T>,"!"]>;
}
type ExpressionPending<T> = DeferredExpression<T>;
type ExpressionReplay = Assert<Equal<ExpressionPending<"value">,"value!">>;
typefn Template<T> {typefn Identity<U> {return U;} return `value:${Identity<T>}`;}
type TemplatePending<T> = Template<T>;
type TemplateReplay = Assert<Equal<TemplatePending<123>,"value:123">>;
typefn Preceding<> {typefn Inner<A,B extends A> {return B;} return Inner<string,"x">;}
type PrecedingConstraint = Assert<Equal<Preceding<>,"x">>;
typefn NestedControl<T> {
 typefn First<U> {return [T,U];}
 typefn Second<U> {if(T == "yes") {return First<U>;} return false;}
 return Second<1>;
}
type NestedControlPending<T> = NestedControl<T>;
type NestedControlYes = Assert<Equal<NestedControlPending<"yes">,["yes",1]>>;
type NestedControlNo = Assert<Equal<NestedControlPending<"no">,false>>;
