package catalog_test

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"testing"
)

func TestDisabledSubmissionsPreserveApprovedHistoryAndTeamReplies(t *testing.T) {
	for _, flags := range []struct {
		name        string
		reviews, qa bool
	}{{"both-disabled", false, false}, {"reviews-only", true, false}, {"qa-only", false, true}, {"both-enabled", true, true}} {
		t.Run(flags.name, func(t *testing.T) {
			c := &catalog.ProductContent{ReviewsEnabled: flags.reviews, QAEnabled: flags.qa}
			p := catalog.ProductInteractions(c, true, catalog.InteractionCapabilities{Reviews: true, QA: true})
			if p.Reviews.NewSubmissions != flags.reviews || p.QA.NewSubmissions != flags.qa {
				t.Fatal("independent submission settings")
			}
			if !p.Reviews.ApprovedContentVisible || !p.QA.ApprovedContentVisible || !p.Reviews.TeamReplies || !p.QA.TeamReplies {
				t.Fatal("disabling submissions hid history or team replies")
			}
		})
	}
}
func TestNewProductInteractionDefaultsAndServiceReadiness(t *testing.T) {
	defaults := catalog.ProductInteractions(nil, true, catalog.InteractionCapabilities{Reviews: true, QA: true})
	if defaults.Reviews.NewSubmissions || defaults.QA.NewSubmissions {
		t.Fatal("new product must default both settings off")
	}
	content := &catalog.ProductContent{ReviewsEnabled: true, QAEnabled: true}
	for _, visible := range []bool{false, true} {
		p := catalog.ProductInteractions(content, visible, catalog.InteractionCapabilities{})
		if p != (catalog.ProductInteractionPolicy{}) {
			t.Fatal("unimplemented service advertised capabilities")
		}
	}
	hidden := catalog.ProductInteractions(content, false, catalog.InteractionCapabilities{Reviews: true, QA: true})
	if hidden != (catalog.ProductInteractionPolicy{}) {
		t.Fatal("hidden product exposed interactions")
	}
	onlyQA := catalog.ProductInteractions(content, true, catalog.InteractionCapabilities{QA: true})
	if onlyQA.Reviews != (catalog.InteractionAccess{}) || !onlyQA.QA.NewSubmissions {
		t.Fatal("service readiness must also be independent")
	}
}
