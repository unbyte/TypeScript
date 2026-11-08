typefn F<T> { for (type item of T) { return item; } return never; } type R = F<string[]>;
