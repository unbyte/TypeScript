typefn F<T>: type { return F<T + 1>; } type R = F<0>;
