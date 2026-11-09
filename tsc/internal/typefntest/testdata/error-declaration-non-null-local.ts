export function local() {
  typefn Maybe<T> { return T|null; }
  function read():Maybe<1> { throw 0; }
  const value:Maybe<1> = read();
  return value!;
}
