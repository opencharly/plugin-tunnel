package tunnelverb

// schema_test.go — the served-schema contract of THIS plugin, proved at the two boundaries
// the contract actually has.
//
// The repo's other gates do not cover this: `charly box validate` checks the candy manifest
// and its `plugin:` block, and `go test` of the exec leg never compiles the schema. A plugin
// whose served schema stopped composing — or whose `#TunnelInput` stopped referencing the
// contract's `#TunnelConfig` — would keep every one of those green while failing at start-up
// on the host. So this file composes the schema the way each half of the contract composes
// it, and judges the SAME authored inputs through both:
//
//   - SERVE side — `NewMeta().Describe` → `sdk.BuildCapabilities` →
//     `schemacompose.ServedSchema`, exactly the call the provider makes when it advertises
//     verb:tunnel;
//   - HOST side — the `Capabilities.schema_cue` body that travelled over the wire, composed
//     by `schemacompose.ServedText` under the DERIVED module declaration (charly's
//     registerPluginUnitSchema calls the same function on that same text).
//
// The input fixture is the equivalence record: the same eleven authored documents the plugin
// accepted and rejected before `#TunnelConfig` / `#TunnelPort` were replaced by the import of
// `github.com/opencharly/spec/schema`, three accepted and eight rejected.

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	"github.com/opencharly/plugin-tunnel/candy/plugin-tunnel/params"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/schemacompose"
	"github.com/opencharly/spec/spec"
)

// The plugin's params must carry the CONTRACT's wire types, not copies of them: this
// assignment fails to compile if the schema ever stops importing schema.#TunnelConfig and
// goes back to declaring a local twin.
var _ spec.TunnelConfig = params.TunnelInput{}.Config

const tunnelInputDef = "#TunnelInput"

// servedUnit composes the plugin's embedded schema the way the SERVE side does.
func servedUnit(t *testing.T) cue.Value {
	t.Helper()
	v, err := schemacompose.ServedSchema(nil, schemaFS, "schema")
	if err != nil {
		t.Fatalf("the serve side cannot compose candy/plugin-tunnel/schema: %v", err)
	}
	return v
}

// hostUnit takes the capabilities the provider actually advertises over Describe, then
// composes the schema body the host receives the way the HOST gate does.
func hostUnit(t *testing.T) (string, cue.Value) {
	t.Helper()
	caps, err := NewMeta().Describe(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if caps == nil {
		t.Fatal("Describe returned no capabilities")
	}
	body := caps.GetSchemaCue()
	if strings.TrimSpace(body) == "" {
		t.Fatal("the plugin advertises an EMPTY schema_cue")
	}
	v, err := schemacompose.ServedText(nil, body)
	if err != nil {
		t.Fatalf("the host side cannot compose the served schema_cue: %v", err)
	}
	return body, v
}

// topLevelDefs returns the top-level definition names a composed unit value declares.
func topLevelDefs(t *testing.T, v cue.Value) map[string]bool {
	t.Helper()
	it, err := v.Fields(cue.Definitions(true))
	if err != nil {
		t.Fatalf("reading the composed unit's definitions: %v", err)
	}
	out := map[string]bool{}
	for it.Next() {
		out[it.Selector().String()] = true
	}
	return out
}

// judge applies the SAME rule the host's validator applies (plugin_loader.go:
// `def.Unify(input).Validate(cue.Concrete(true))`) and reports "<nil>" for an accepted
// document.
func judge(t *testing.T, unit cue.Value, doc string) string {
	t.Helper()
	d := unit.LookupPath(cue.ParsePath(tunnelInputDef))
	if d.Err() != nil {
		t.Fatalf("composed unit has no %s: %v", tunnelInputDef, d.Err())
	}
	in := cuecontext.New().CompileBytes([]byte(doc))
	if in.Err() != nil {
		t.Fatalf("fixture document does not compile: %v\n%s", in.Err(), doc)
	}
	if err := d.Unify(in).Validate(cue.Concrete(true)); err != nil {
		return err.Error()
	}
	return "<nil>"
}

// tunnelInputFixture is the equivalence record: every authored shape the plugin's input def
// accepted or rejected before the restated #TunnelConfig / #TunnelPort were deleted.
var tunnelInputFixture = []struct {
	name  string
	doc   string
	valid bool
}{
	{
		name:  "plan step with config and expect",
		doc:   `{"method":"plan","config":{"provider":"tailscale","box_name":"plugin-tunnel","ports":[{"port":8888,"backend_port":8888,"protocol":"http","public":false}]},"expect":["tailscale serve --bg --https=8888 http://127.0.0.1:8888"]}`,
		valid: true,
	},
	{
		name:  "start with an empty config",
		doc:   `{"method":"start","config":{}}`,
		valid: true,
	},
	{
		name:  "stop without a config",
		doc:   `{"method":"stop"}`,
		valid: true,
	},
	{
		name:  "empty method",
		doc:   `{"method":""}`,
		valid: false,
	},
	{
		name:  "missing method",
		doc:   `{}`,
		valid: false,
	},
	{
		name:  "string port",
		doc:   `{"method":"start","config":{"ports":[{"port":"8888"}]}}`,
		valid: false,
	},
	{
		name:  "unknown key at the input level",
		doc:   `{"method":"start","bogus":1}`,
		valid: false,
	},
	{
		name:  "unknown key at the config level",
		doc:   `{"method":"start","config":{"bogus":1}}`,
		valid: false,
	},
	{
		name:  "unknown key at the port level",
		doc:   `{"method":"start","config":{"ports":[{"bogus":1}]}}`,
		valid: false,
	},
	{
		name:  "string expect",
		doc:   `{"method":"plan","expect":"tailscale serve --bg --https=8888 http://127.0.0.1:8888"}`,
		valid: false,
	},
	{
		name:  "string public",
		doc:   `{"method":"start","config":{"ports":[{"public":"true"}]}}`,
		valid: false,
	},
}

// TestServedSchemaLoadsAndCarriesTheImport is the load proof R7 asks for at this boundary:
// the schema COMPOSES as one CUE module instance (the serve-side compose fails loudly if the
// import cannot resolve), the wire body carries that import, and the only definition this
// plugin declares is its own #TunnelInput — the restated #TunnelConfig / #TunnelPort are gone.
func TestServedSchemaLoadsAndCarriesTheImport(t *testing.T) {
	serveDefs := topLevelDefs(t, servedUnit(t))
	if !serveDefs[tunnelInputDef] {
		t.Fatalf("the served schema declares no %s: %v", tunnelInputDef, serveDefs)
	}
	for _, deleted := range []string{"#TunnelConfig", "#TunnelPort"} {
		if serveDefs[deleted] {
			t.Errorf("the served schema still DECLARES %s — the restated twin must be deleted, not kept beside the import", deleted)
		}
	}

	body, host := hostUnit(t)
	if !strings.Contains(body, `import "github.com/opencharly/spec/schema"`) {
		t.Errorf("the served schema_cue does not import the contract module:\n%s", body)
	}
	hostDefs := topLevelDefs(t, host)
	if !hostDefs[tunnelInputDef] {
		t.Fatalf("the host-composed unit declares no %s: %v", tunnelInputDef, hostDefs)
	}
	for _, deleted := range []string{"#TunnelConfig", "#TunnelPort"} {
		if hostDefs[deleted] {
			t.Errorf("the host-composed unit DECLARES %s — the wire body must carry only the plugin's own defs", deleted)
		}
	}
}

// TestTunnelInputFixtureJudgedIdenticallyByBothHalves is the equivalence proof: the eleven
// authored documents are judged by the value the serve side composes AND by the value the
// host composes from the wire, both must agree, and both must agree with the verdict the
// plugin gave before the cutover.
func TestTunnelInputFixtureJudgedIdenticallyByBothHalves(t *testing.T) {
	serve := servedUnit(t)
	_, host := hostUnit(t)

	accepted, rejected := 0, 0
	for _, tc := range tunnelInputFixture {
		gotServe := judge(t, serve, tc.doc)
		gotHost := judge(t, host, tc.doc)
		t.Logf("%-40s accept=%v verdict=%s", tc.name, tc.valid, gotServe)
		if gotServe != gotHost {
			t.Errorf("%s: the two halves disagree\n serve: %s\n host:  %s\n doc: %s", tc.name, gotServe, gotHost, tc.doc)
		}
		if tc.valid {
			accepted++
			if gotServe != "<nil>" {
				t.Errorf("%s must be ACCEPTED, got: %s", tc.name, gotServe)
			}
			continue
		}
		rejected++
		if gotServe == "<nil>" {
			t.Errorf("%s must be REJECTED, but the schema accepted it", tc.name)
		}
	}
	if accepted != 3 || rejected != 8 {
		t.Fatalf("the fixture must keep its measured shape (3 accept / 8 reject), got %d/%d", accepted, rejected)
	}
	t.Logf("fixture: %d accepted, %d rejected, both halves identical", accepted, rejected)
}

// TestImportIsEnforcedByTheImportedDefinition proves the import is LIVE rather than
// decorative: a document that violates the contract's own constraint (`port` is an int there)
// is rejected with the imported constraint's error, while a conforming one passes.
func TestImportIsEnforcedByTheImportedDefinition(t *testing.T) {
	serve := servedUnit(t)
	if got := judge(t, serve, `{"method":"start","config":{"ports":[{"port":true}]}}`); got == "<nil>" {
		t.Fatal("a config violating the IMPORTED spec.#TunnelPort must be rejected")
	}
	if got := judge(t, serve, `{"method":"start","config":{"provider":"cloudflare","ports":[{"port":8080}]}}`); got != "<nil>" {
		t.Fatalf("a conforming config must pass: %s", got)
	}
}

// TestParamsConfigIsTheContractType states the same relation in one place a reader can read
// without a compiler: the generated field is spec.TunnelConfig, and its JSON key is unchanged.
func TestParamsConfigIsTheContractType(t *testing.T) {
	var in params.TunnelInput
	var cfg spec.TunnelConfig
	in.Config = cfg
	if fmt.Sprintf("%T", in.Config) != "spec.TunnelConfig" {
		t.Fatalf("params.TunnelInput.Config is %T, want spec.TunnelConfig", in.Config)
	}
}
