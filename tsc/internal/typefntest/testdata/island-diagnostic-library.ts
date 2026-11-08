type Constrained<T extends string> = { value: T };
export typefn Make<T> { return type(Constrained<T>); }
