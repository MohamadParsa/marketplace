package backofficememory

import (
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/finance"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"github.com/cherriva/marketplace/apps/api/internal/domain/support"
)

func Fixture() backoffice.Workspace {
	state := backoffice.Workspace{
		Users:      []backoffice.User{{ID: "admin", Phone: "09000000001", Name: "مدیر آزمایشی", Role: identity.RoleSuperAdmin, Active: true}, {ID: "ops", Phone: "09000000002", Name: "عملیات آزمایشی", Role: identity.RoleOperations, Active: true}, {ID: "support", Phone: "09000000003", Name: "پشتیبان آزمایشی", Role: identity.RoleSupport, Active: true}, {ID: "finance", Phone: "09000000004", Name: "مالی آزمایشی", Role: identity.RoleFinance, Active: true}, {ID: "viewer", Phone: "09000000005", Name: "مشاهده‌گر آزمایشی", Role: identity.RoleViewer, Active: true}},
		Stores:     []store.Store{{ID: "s1", Name: "خانه سبز", Owner: "فروشنده نمونه الف", City: "تهران", Active: true, Model: "shared"}, {ID: "s2", Name: "چایستان", Owner: "فروشنده نمونه ب", City: "رشت", Active: true, Model: "independent"}, {ID: "s3", Name: "دیجیتال آبان", Owner: "فروشنده نمونه ج", City: "شیراز", Active: false, Model: "shared"}, {ID: "s4", Name: "استودیو آرا", Owner: "فروشنده نمونه د", City: "اصفهان", Active: true, Model: "independent"}},
		Brands:     []catalog.Brand{{ID: "b1", Name: "اپل", Active: true, Version: 1}, {ID: "b2", Name: "سامسونگ", Active: true, Version: 1}, {ID: "b3", Name: "شیائومی", Active: true, Version: 1}},
		Categories: []catalog.Category{{ID: "c1", Name: "خانه و زندگی"}, {ID: "c2", Name: "چای و دمنوش"}, {ID: "c3", Name: "کالای دیجیتال"}},
		Products:   []catalog.Product{{ID: "p1", Name: "چراغ رومیزی مینیمال", CategoryID: "c1", Model: "shared", Status: "published", Version: 1}, {ID: "p2", Name: "چای بهاره ۵۰۰ گرمی", CategoryID: "c2", Model: "independent", Status: "pending", Version: 1}, {ID: "p3", Name: "هدفون بی‌سیم", CategoryID: "c3", Model: "shared", Status: "pending", Version: 1}, {ID: "p4", Name: "گلدان سرامیکی", CategoryID: "c1", Model: "independent", Status: "published", Version: 1}, {ID: "p5", Name: "دمنوش به‌لیمو", CategoryID: "c2", Model: "independent", Status: "rejected", Version: 1}},
		Offers:     []catalog.Offer{{ID: "of1", ProductID: "p1", StoreID: "s1", Stock: 3, Price: 18500000}, {ID: "of2", ProductID: "p2", StoreID: "s2", Stock: 24, Price: 4200000}, {ID: "of3", ProductID: "p3", StoreID: "s3", Stock: 8, Price: 32000000}, {ID: "of4", ProductID: "p4", StoreID: "s4", Stock: 2, Price: 6800000}},
		Orders: []orders.Order{
			{ID: "ORD-1042", PurchaseID: "BUY-501", StoreID: "s1", City: "تهران", Status: "on-hold", Payment: "success", CreatedAt: "2026-11-12T08:00:00Z", Items: []orders.Item{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 18500000, Status: "supply-exception"}}, Version: 1, Packages: []string{}},
			{ID: "ORD-1043", PurchaseID: "BUY-501", StoreID: "s2", City: "تهران", Status: "preparing", Payment: "success", CreatedAt: "2026-11-12T08:00:00Z", Items: []orders.Item{{ID: "i2", ProductID: "p2", Quantity: 2, UnitPrice: 4200000, Status: "confirmed"}}, Version: 1, Packages: []string{}},
			{ID: "ORD-1044", PurchaseID: "BUY-502", StoreID: "s4", City: "تبریز", Status: "shipped", Payment: "success", CreatedAt: "2026-11-11T10:00:00Z", Items: []orders.Item{{ID: "i3", ProductID: "p4", Quantity: 2, UnitPrice: 6800000, Status: "shipped"}}, Version: 1, Packages: []string{"TEST-PACK-01", "TEST-PACK-02"}},
			{ID: "ORD-1045", PurchaseID: "BUY-503", StoreID: "s1", City: "کرج", Status: "awaiting-payment", Payment: "unknown", CreatedAt: "2026-11-12T09:00:00Z", Items: []orders.Item{{ID: "i4", ProductID: "p1", Quantity: 1, UnitPrice: 18500000, Status: "awaiting-confirmation"}}, Version: 1, Packages: []string{}},
		},
		Cases:         []support.Case{{ID: "CASE-201", OrderID: "ORD-1042", ItemID: "i1", Kind: "عدم تأمین", Status: "open", Assignee: "support", Reason: "کمبود موجودی فروشنده", Deadline: "2026-11-14T20:29:59Z", Version: 1}, {ID: "CASE-202", OrderID: "ORD-1044", ItemID: "i3", Kind: "کسری بسته", Status: "in-progress", Assignee: "support", Reason: "گزارش آزمایشی کسری یکی از بسته‌ها", Deadline: "2026-11-14T20:29:59Z", Version: 1}},
		Assignments:   []support.Assignment{},
		Transfers:     []finance.Transfer{{ID: "TX-301", Kind: "payment", OrderID: "ORD-1043", ItemID: "i2", StoreID: "s2", Amount: 8400000, Status: "success", Source: "TEST-CARD-A", Reference: "TEST-PAY-301", Version: 1}, {ID: "TX-302", Kind: "refund", OrderID: "ORD-1044", ItemID: "i3", StoreID: "s4", Amount: 6800000, Status: "pending", Reserved: true, Source: "TEST-CARD-B", Version: 1}, {ID: "TX-303", Kind: "withdrawal", Amount: 12000000, Status: "pending", Reserved: true, Source: "TEST-WALLET-C", Version: 1}, {ID: "TX-304", Kind: "settlement", OrderID: "ORD-1043", ItemID: "i2", StoreID: "s2", Amount: 4000000, Status: "pending", Reserved: true, Source: "TEST-SELLER-S2", Version: 1}, {ID: "TX-305", Kind: "payment", OrderID: "ORD-1045", ItemID: "i4", StoreID: "s1", Amount: 18500000, Status: "unknown", Reserved: true, Source: "TEST-CARD-D", Version: 1}, {ID: "TX-306", Kind: "debt", OrderID: "ORD-1044", ItemID: "i3", StoreID: "s4", Amount: 6800000, Status: "outstanding", Source: "TEST-REFUND-OLD", Version: 1}},
		Ledger:        []finance.Entry{},
		Rules:         []backoffice.Rule{{ID: "rule1", Name: "کمیسیون خانه و زندگی", Level: "category", Kind: "percent", Value: 5, Active: true}, {ID: "rule2", Name: "قرارداد صفر چایستان", Level: "store", Kind: "percent", Value: 0, Active: true}},
		Shipping:      []backoffice.Shipping{{ID: "ship1", Method: "پست پیشتاز", City: "تهران", Tariff: 650000, Days: 2}, {ID: "ship2", Method: "پیک شهری", City: "تهران", Tariff: 950000, Days: 1}, {ID: "ship3", Method: "پست پیشتاز", City: "تبریز", Tariff: 850000, Days: 3}},
		Notifications: []backoffice.Notification{{ID: "n1", Event: "ارجاع عدم تأمین ORD-1042", Recipient: "پشتیبانی", Status: "failure", Retries: 1}, {ID: "n2", Event: "ثبت خرید BUY-501", Recipient: "مشتری آزمایشی", Status: "success", Retries: 0}, {ID: "n3", Event: "ارسال ORD-1044", Recipient: "مشتری آزمایشی", Status: "pending", Retries: 0}},
		Audit:         []backoffice.Audit{},
	}
	fixtureAttributes(&state)
	return state
}
