import type {Double} from './out/main';
type Assert<T extends true>=T;
type Direct=Assert<Double<21> extends 42 ? true : false>;
type Generic<T>=Double<T>;
type Replay=Assert<Generic<22> extends 44 ? true : false>;
