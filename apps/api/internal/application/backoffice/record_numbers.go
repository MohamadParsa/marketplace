package backoffice

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Display numbers are an additive registry. Internal keys and historical references never change.
func recordKeys(w Workspace) []string {
	keys := []string{}
	add := func(kind string, rows any) {
		v := reflect.ValueOf(rows)
		for i := 0; i < v.Len(); i++ {
			id := v.Index(i).FieldByName("ID")
			if id.IsValid() && id.String() != "" {
				keys = append(keys, kind+"|"+id.String())
			}
		}
	}
	add("product", w.Products)
	add("variant", w.Variants)
	add("offer", w.Offers)
	add("category", w.Categories)
	add("attribute", w.Attributes)
	add("brand", w.Brands)
	add("store", w.Stores)
	add("user", w.Users)
	add("order", w.Orders)
	add("case", w.Cases)
	add("transfer", w.Transfers)
	add("ledger", w.Ledger)
	add("notification", w.Notifications)
	add("banner", w.Banners)
	add("assignment", w.Assignments)
	if w.Access != nil {
		add("role", w.Access.Roles)
	}
	sort.Strings(keys)
	return keys
}

func EnsureRecordNumbers(w *Workspace) {
	if w.RecordNumbers == nil {
		w.RecordNumbers = map[string]int64{}
	}
	for _, n := range w.RecordNumbers {
		if n > w.LastRecordNumber {
			w.LastRecordNumber = n
		}
	}
	variantNumbers := map[int64]string{}
	for key, number := range w.RecordNumbers {
		if strings.HasPrefix(key, "variant|") {
			variantNumbers[number] = key
		}
	}
	for _, v := range w.Variants {
		if strings.HasPrefix(v.SKU, "chr-") {
			n, e := strconv.ParseInt(strings.TrimPrefix(v.SKU, "chr-"), 10, 64)
			if e == nil && n > 0 {
				key := "variant|" + v.ID
				if owner := variantNumbers[n]; w.RecordNumbers[key] == 0 && (owner == "" || owner == key) {
					w.RecordNumbers[key] = n
					variantNumbers[n] = key
				}
				if n > w.LastRecordNumber {
					w.LastRecordNumber = n
				}
			}
		}
	}
	for _, key := range recordKeys(*w) {
		if w.RecordNumbers[key] == 0 {
			w.LastRecordNumber++
			w.RecordNumbers[key] = w.LastRecordNumber
		}
	}
}

func recordNumber(w *Workspace, kind, id string) int64 {
	EnsureRecordNumbers(w)
	key := kind + "|" + id
	if w.RecordNumbers[key] == 0 {
		w.LastRecordNumber++
		w.RecordNumbers[key] = w.LastRecordNumber
	}
	return w.RecordNumbers[key]
}

func recordCode(w *Workspace, kind, id string) string {
	return "chr-" + strconv.FormatInt(recordNumber(w, kind, id), 10)
}
func variantCode(w *Workspace, id string) string { return recordCode(w, "variant", id) }

func scopedRecordNumbers(w Workspace, all map[string]int64) map[string]int64 {
	out := map[string]int64{}
	for _, key := range recordKeys(w) {
		if number := all[key]; number > 0 {
			out[key] = number
		}
	}
	return out
}
