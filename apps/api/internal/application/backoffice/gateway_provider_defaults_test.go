package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"reflect"
	"strings"
	"testing"
)

func TestEverySupportedGatewayAlwaysHasSettingsAndSingleSwitchableRecord(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	before, _ := h.repo.Snapshot(ctx)
	dir, err := h.s.ReadGatewaySettings(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(dir.Items) != len(dir.Providers) || len(dir.Items) != 1 || dir.Items[0].ID != "" || dir.Items[0].Active || dir.Items[0].ExecutionReady {
		t.Fatal("missing default provider settings")
	}
	for _, p := range dir.Providers {
		for _, section := range p.Sections {
			if section.Expanded {
				t.Fatal("settings section expanded initially")
			}
		}
	}
	afterRead, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Commerce, afterRead.Commerce) {
		t.Fatal("reading defaults wrote state")
	}
	in := dir.Items[0].GatewaySettingsInput
	in.Active = true
	raw, err := h.s.SaveGatewaySettings(ctx, h.admin, in, h.key())
	saved := decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if !saved.Active || saved.ExecutionReady || saved.ID == "" {
		t.Fatal("settings switch failed")
	}
	if _, err = h.s.SaveGatewaySettings(ctx, h.admin, in, h.key()); err != backoffice.ErrConflict {
		t.Fatal("duplicate provider instance allowed", err)
	}
	dir, err = h.s.ReadGatewaySettings(ctx, h.admin)
	if err != nil || len(dir.Items) != 1 || !dir.Items[0].Active {
		t.Fatal("read switched state", err)
	}
	edit := saved.GatewaySettingsInput
	edit.Active = false
	raw, err = h.s.SaveGatewaySettings(ctx, h.admin, edit, h.key())
	saved = decodeCommerce[backoffice.GatewaySettingsView](t, raw, err)
	if saved.Active || saved.Version != 2 {
		t.Fatal("switch off failed")
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.Commerce.GatewaySettings) != 1 || !reflect.DeepEqual(before.Commerce.Settings.Methods, state.Commerce.Settings.Methods) || !reflect.DeepEqual(before.Commerce.Entries, state.Commerce.Entries) {
		t.Fatal("switch changed payment availability or money")
	}
	public, err := h.s.PublicCommerceConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(public)
	if strings.Contains(string(body), "zarinpal") {
		t.Fatal("unimplemented gateway exposed")
	}
}
