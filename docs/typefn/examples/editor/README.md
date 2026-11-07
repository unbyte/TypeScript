# Runnable editor example

Open this directory with the fork's native language server. Follow
[the editor setup](../../editor-integration.md) and enable
`js/ts.inlayHints.typefnResults.enabled` for result overlays.

Try hover, definition, references, and rename on `Twice` in `index.ts`. In
`types.ts`, try completion and rename on `result` and `item`. Edit `Twice` to
return `result + N`: the consumer's inferred result changes from 42 to 63 and
its assignment receives a native TypeScript diagnostic. Restore it afterward.

`Deferred<N>` keeps the addition pending; `Deferred<21>` specializes it to
42. Hover on its generic `Twice<N>` application shows the signature without a
completed-result overlay. Changing the specialization to `Deferred<"x">` produces
an arithmetic diagnostic. Restore `21` afterward.

Validate from the repository root:

```sh
built/local/tsgo --project docs/typefn/examples/editor
```

These examples use the implemented execution subset; they are checked by the fork.
