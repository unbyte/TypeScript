import { TypeFlags } from "@typescript/typescript/sync";
import assert from "node:assert/strict";
import { test } from "node:test";

test("deferred type flags retain their identity across unsigned wire values and JavaScript masks", () => {
    const wireFlags = TypeFlags.Typefn >>> 0;
    assert.equal(wireFlags, TypeFlags.Typefn);
    assert.equal(wireFlags & TypeFlags.Instantiable, TypeFlags.Typefn);
    assert.equal(wireFlags & TypeFlags.StructuredOrInstantiable, TypeFlags.Typefn);
    assert.equal(wireFlags & TypeFlags.AnyOrUnknown, 0);
});
