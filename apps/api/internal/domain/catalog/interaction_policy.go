package catalog

// InteractionCapabilities is set by an implemented service, not by product toggles.
type InteractionCapabilities struct {
	Reviews bool
	QA      bool
}

type InteractionAccess struct {
	NewSubmissions         bool
	ApprovedContentVisible bool
	// TeamReplies is a feature gate; a handler must also authenticate and authorize staff.
	TeamReplies bool
}
type ProductInteractionPolicy struct {
	Reviews InteractionAccess
	QA      InteractionAccess
}

// ProductInteractions implements the owner-confirmed R2-Q09 policy. Turning off
// submissions never deletes/hides approved history or prevents authorized team
// replies. Unpublished content and unavailable services grant no capabilities.
func ProductInteractions(content *ProductContent, publiclyVisible bool, ready InteractionCapabilities) ProductInteractionPolicy {
	reviews, qa := false, false
	if content != nil {
		reviews, qa = content.ReviewsEnabled, content.QAEnabled
	}
	access := func(enabled, available bool) InteractionAccess {
		available = publiclyVisible && available
		return InteractionAccess{NewSubmissions: available && enabled, ApprovedContentVisible: available, TeamReplies: available}
	}
	return ProductInteractionPolicy{Reviews: access(reviews, ready.Reviews), QA: access(qa, ready.QA)}
}
