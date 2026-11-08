import type * as Helpers from './provider';
export typefn Public<T> {
 type value = Helpers.Visible<T>;
 return Helpers.Hidden<value>;
}
