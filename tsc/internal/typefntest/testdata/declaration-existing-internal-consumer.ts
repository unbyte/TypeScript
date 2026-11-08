import type {Public} from './out/library';
type Assert<T extends true>=T;
type Check=Assert<Public<1> extends 1 ? true : false>;
