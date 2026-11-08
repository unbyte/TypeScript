type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Keys<> { return { 0x10: "hex"; 1e3: "exponent"; 1_234: "separator"; \u0061: "escaped"; type: "keyword" }; }
type Check = Assert<Equal<Keys<>, {16: "hex"; 1000: "exponent"; 1234: "separator"; a: "escaped"; type: "keyword"}>>;
typefn SameName<K> { return { [K]: 1 }; }
type Computed = Assert<Equal<SameName<16>, {16: 1}>>;
