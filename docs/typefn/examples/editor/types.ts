export interface Box<T> {
  value: T;
}

/** Compute a numeric literal through mutable local Types. */
export typefn Twice<N> {
  type result = N;
  result = result + N;
  return result;
}

/** Build a tuple, preserving the native Type of each element. */
export typefn Reverse<Items> {
  type result = [];
  for (type item of Items) {
    result = [item, ...result];
  }
  return result;
}

/** Compose native generic Types inside a type function. */
export typefn FrozenBox<T> {
  return Readonly<Box<T>>;
}
