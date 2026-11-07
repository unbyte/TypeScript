type Box<T extends string> = [T]; typefn F<T> { return Box<T>; } type R = F<number>;
