typefn F<T> { return T["missing"]; } type R = F<{key: 1}>;
