---
sidebar_position: 9
---

# Stability

From **v1.0.0**, apiary follows [Semantic Versioning](https://semver.org). Within
the v1 series the **annotation format**, **struct tags** (and their constraint
mapping), **supported handler signatures**, and **CLI flags / config keys** will
not change in a breaking way.

Not covered by the guarantee: the exact byte layout of generated YAML/JSON (only
its validity as an OpenAPI 3.1 document and the documented semantics), diagnostic
wording, internal packages, and behavior for input that does not compile.

Deprecated features are announced in the
[changelog](https://github.com/yaop-labs/apiary/blob/main/CHANGELOG.md), keep
working for the remainder of v1, and are removed no earlier than v2. See
[`STABILITY.md`](https://github.com/yaop-labs/apiary/blob/main/STABILITY.md) for
the full policy.
