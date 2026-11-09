# ops branch

Upstream 3x-ui v3.10.2 (608ef4e) plus the two patches in `ops/patches/`, built for linux/amd64 by
`.github/workflows/ops-release.yml` when a `v3.10.2-ops.N` tag is pushed. The release carries the static panel
binary, `manifest.json` (patch and binary SHA-256, toolchain, build run) and `SHA256SUMS`.

Servers install a release by its pinned SHA-256; nothing here is fetched unpinned. Upstream's own workflows are
removed on this branch so that tags here never publish upstream's release or Docker image.
