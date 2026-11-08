interface Animal { animal: true; }
interface Dog extends Animal { dog: true; }
typefn Fn<T> { return types.function<[T], void>; }
declare const dog: Fn<Dog>;
const animal: Fn<Animal> = dog;
