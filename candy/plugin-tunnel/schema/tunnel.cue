// The OUT-OF-TREE plugin-tunnel's OWN CUE schema — the typed params for the
// `tunnel` VERB (verb:tunnel), the externalized tailscale/cloudflare TUNNEL
// EXECUTION LEG. It is the SINGLE SOURCE for this plugin's params, used two ways
// (the same contract the core `spec` module + the reference examplerunverb use):
//
//  1. GENERATE the Go param struct — the schema→Go pipeline (`spec/schemaparams.Generate`,
//     reached from a shell as `charly candy params plugin-tunnel`) concatenates this file
//     under a `package params` + `@go(params)` header, composes it as ONE offline CUE module
//     instance, and emits ../params/cue_types_gen.go, so the provider decodes plugin_input
//     into a TYPED struct (params.TunnelInput, whose Config is spec.TunnelConfig), never a
//     hand-parsed map.
//  2. VALIDATE authored input AT RUNTIME — the plugin SERVES this source over the
//     Describe channel (sdk.BuildCapabilities → schema_cue); the host composes it per unit
//     (spec/schemacompose.ServedText) and unifies it onto its base, then validates every
//     authored `plugin: tunnel` step's plugin_input against #TunnelInput. The host NEVER reads
//     this file from disk — the schema travels with the plugin.
//
// ONE NAME, ONE HOME. #TunnelConfig / #TunnelPort are the CONTRACT module's definitions
// (`github.com/opencharly/spec/schema`): the resolved WIRE carrier spec.TunnelConfig /
// spec.TunnelPort that the pod-lifecycle plugins (candy/plugin-deploy-pod for start/stop,
// candy/plugin-pod for remove) marshal over this verb's {method,config} Invoke envelope, and
// that sdk/deploykit's quadlet emitter consumes. This schema IMPORTS them; the structurally
// identical local twins it used to carry are DELETED (R2 — one name, one home). #TunnelInput
// is this plugin's OWN authoring shape and stays here: it is the def the plugin advertises as
// its InputDef, and the only def the host's input validator is keyed on.
//
// verb:tunnel is NOT a check verb in the usual sense — it is the externalized TUNNEL
// EXECUTION backend. The pod-lifecycle plugins drive its start/stop/setup methods directly
// over this verb's Invoke envelope ({method,config}) via InvokeProvider. It ALSO carries a
// benign `plan` (dry-run) method — a `plugin: tunnel` check step returns the EXACT
// tailscale/cloudflared argv it WOULD run WITHOUT exec, so a disposable bed proves the
// registry dispatch + the TunnelConfig wire round-trip + the moved command-building logic with
// ZERO tailscale/cloudflare credentials.
//
// THE IMPORT RESOLVES OFFLINE. It is the shape spec/schemacompose exists for: the contract
// module is served from the spec Go module the binary already links — no network, no module
// cache, no `cue.mod`, no published CUE module — on the serve side AND on the host side, which
// composes the SAME body under the DERIVED module declaration (the wire carries no module
// declaration, so both sides must evaluate under one they can derive).

import "github.com/opencharly/spec/schema"

// #TunnelInput — the plugin_input shape for the `tunnel` verb.
#TunnelInput: {
	// method — the tunnel operation: start | stop | setup | plan. `plan` is the
	// creds-free dry-run that returns the argv the operation WOULD run (no exec).
	method: string & !="" @go(Method)
	// config — the resolved tunnel configuration to act on (the contract's TunnelConfig,
	// sent by the calling pod-lifecycle plugin over the Invoke envelope).
	config?: schema.#TunnelConfig @go(Config)
	// expect — plan only: the expected argv command lines (space-joined). When set, the
	// plan method compares the built argv against it (FAIL on mismatch); empty ⇒ the
	// step passes as long as dispatch + round-trip succeed, echoing the built argv.
	expect?: [...string] @go(Expect)
}
