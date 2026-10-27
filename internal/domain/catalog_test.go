package domain

import "testing"

func TestPublishedRBACPolicy(t *testing.T) {
	policy, err := NewPolicy()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		role       Role
		permission Permission
		allowed    bool
	}{
		{Member, PurchaseCreate, true},
		{Member, PurchaseDecide, false},
		{Member, MembershipManage, false},
		{Approver, PurchaseCreate, true},
		{Approver, AuditRead, true},
		{Approver, PurchaseDecide, true},
		{Approver, MembershipManage, false},
		{Administrator, PurchaseCreate, true},
		{Administrator, PurchaseDecide, true},
		{Administrator, MembershipManage, true},
		{"unknown", PurchaseCreate, false},
		{Administrator, "unknown", false},
	} {
		principal := Principal{
			ActorID:  "actor",
			TenantID: "tenant",
			Role:     test.role,
		}
		if got := principal.Can(policy, test.permission); got != test.allowed {
			t.Errorf(
				"role=%s permission=%s allowed=%v",
				test.role,
				test.permission,
				got,
			)
		}

		if principal.Can(nil, test.permission) {
			t.Error("nil policy retained a grant")
		}
	}

	if !ValidRole(policy, Administrator) || ValidRole(policy, "unknown") ||
		ValidRole(nil, Administrator) {
		t.Fatal("role validation diverged from the compiled policy")
	}

	for _, principal := range []Principal{
		{TenantID: "tenant", Role: Administrator},
		{ActorID: "actor", Role: Administrator},
	} {
		if principal.Can(policy, MembershipManage) {
			t.Fatal("unverified principal retained a capability")
		}
	}
}
