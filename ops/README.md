# ops branch

Upstream 3x-ui v3.10.2 (608ef4e) plus the three patches in `ops/patches/`, built for linux/amd64 by
`.github/workflows/ops-release.yml` when a `v3.10.2-ops.N` tag is pushed. The release carries the static panel
binary, `manifest.json` (patch and binary SHA-256, toolchain, build run) and `SHA256SUMS`.

Servers install a release by its pinned SHA-256; nothing here is fetched unpinned. Upstream's own workflows are
removed on this branch so that tags here never publish upstream's release or Docker image.

`v3.10.2-ops.4` keeps the ops.3 upstream base and billing behavior. Built-in panel/node updates,
Xray installation, Geo downloads and update-channel writes are refused by the API and service layer.
Native Xray `geodata.assets` cannot be saved; older/restored templates keep their stored contents but
the downloader section is omitted from generated runtime configs. Geo browsing and ordinary config
application, start/stop and restart remain available. The UI replaces update controls with an operations
notice. This policy has no runtime opt-out.

The manifest adds `managed_updates_disabled`, `disabled_operations`, `existing_geodata_policy`,
`managed_updates_patch_sha256`, `test_packages` and `frontend_tests`. The release job reconstructs the
source from upstream plus all three patches, runs the listed Go/frontend tests, and publishes a
candidate binary and manifest for the pinned operations workflow. Publishing does not deploy a server.
