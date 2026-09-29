# plugin-tunnel

Tunnel execution for OpenCharly — the `tunnel:` verb.

`plugin-tunnel` is the externalized tailscale/cloudflare **execution leg**. It
runs the actual `tailscale serve`/`funnel` and `cloudflared` lifecycle, stopping
at the exec/auth boundary. The pod-lifecycle plugins that resolve a
`TunnelConfig` (for pod start/stop/remove) drive this verb's
`start`/`stop`/`setup` methods directly over `InvokeProvider`.

The provider is **dual-placement**: compiled into `charly` when listed in
`charly.yml` `compiled_plugins:` (the default), or served out-of-process over
go-plugin gRPC by the `cmd/serve` shim. Placement is invisible above the provider
registry. The **resolution** half of the tunnel subsystem lives in the SDK
(`sdk/deploykit/tunnel_resolve.go`); only the execution leg lives here.

`verb:tunnel` also carries a benign `plan` (dry-run) method: given a
`TunnelConfig` it returns the exact `tailscale`/`cloudflared` argv it **would**
run without exec — so a disposable bed can prove the dispatch and the argv
round-trip with zero tailscale/cloudflare credentials.

## What it provides

| Capability | Surface |
|---|---|
| `verb:tunnel` | the tunnel execution leg — `start`, `stop`, `setup`, and the creds-free `plan` dry-run |

## How to use it

Compose the plugin candy where a box or deploy declares a tunnel:

```yaml
- '@github.com/opencharly/plugin-tunnel/candy/plugin-tunnel:<tag>'
```

The R10 consumer is `box/fedora`'s `check-tunnel-pod` bed, which authors a
`tunnel: {method: plan, config: …}` step asserting the built argv.

## Layout

- `candy/plugin-tunnel/` — the plugin module: `plugin.go` (the provider +
  `NewMeta()`), `tunnel_exec.go` (the execution leg), `schema/tunnel.cue`,
  `params/cue_types_gen.go`, `cmd/serve/main.go`.
- `charly.yml` — the root project manifest (`discover: candy`).
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: `/charly-core:deploy` — the deploy surface whose
  `TunnelConfig`/tunnel lifecycle this verb executes. This candy carries no
  `skill:` entity of its own; the gap is tracked in
  [opencharly/opencharly#291](https://github.com/opencharly/opencharly/issues/291).
- `/charly-internals:plugin` — the plugin/provider model.
- [`opencharly/charly`](https://github.com/opencharly/charly) — the charly CLI.
