type Left = { right: Right };
type Right = { left: Left };
export typefn Pair<T> { return [T, Left]; }
