export function outer<T extends object>() {
  typefn Count<U extends object> { return tuple.length<reflect.properties<U>>; }
  function read(): Count<T> { throw 0; }
  return { value: read() };
}
