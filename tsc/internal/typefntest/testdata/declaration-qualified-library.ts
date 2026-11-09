import * as api from "./provider";
export const direct = api.read<api.Open>();
export const nested = api.Nested.read<api.Open>();
export const forwarded = api.read;
export const specialized = api.read<api.Open>;
export const stored = specialized();
export const immediate = (api.read<api.Open>)();
namespace Local { export const read = api.read; }
export const local = Local.read<api.Open>();
