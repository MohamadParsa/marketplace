package backoffice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
)

const MaxGeographySeedBytes = 2 << 20

type GeographySeedSource struct {
	Repository    string `json:"repository"`
	Commit        string `json:"commit"`
	Version       string `json:"version"`
	OfficialYear  int    `json:"official_year"`
	CityDataset   string `json:"city_dataset"`
	ProvinceCount int    `json:"province_count"`
	CityCount     int    `json:"city_count"`
}

type GeographySeed struct {
	DatasetID string              `json:"dataset_id"`
	Source    GeographySeedSource `json:"source"`
	Provinces []commerce.Province `json:"provinces"`
}

type GeographyImportResult struct {
	DatasetID        string `json:"dataset_id"`
	ProvincesAdded   int    `json:"provinces_added"`
	CitiesAdded      int    `json:"cities_added"`
	ProvincesMatched int    `json:"provinces_matched"`
	CitiesMatched    int    `json:"cities_matched"`
	TotalProvinces   int    `json:"total_provinces"`
	TotalCities      int    `json:"total_cities"`
	SettingsVersion  int    `json:"settings_version"`
	AlreadyApplied   bool   `json:"already_applied"`
}

var geographyProvinceID = regexp.MustCompile(`^ir-p-[0-9]+$`)
var geographyCityID = regexp.MustCompile(`^ir-c-[0-9]+$`)
var geographySourceCommit = regexp.MustCompile(`^[a-f0-9]{40}$`)

func DecodeGeographySeed(raw []byte) (GeographySeed, error) {
	var seed GeographySeed
	if len(raw) > MaxGeographySeedBytes {
		return seed, fmt.Errorf("geography seed is too large")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&seed); err != nil {
		return seed, fmt.Errorf("invalid geography seed JSON")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return seed, fmt.Errorf("unexpected content after geography seed")
	}
	return seed, ValidateGeographySeed(seed)
}

// فقط تفاوت‌های نگارشی فارسی و فاصله را برای تطبیق رکوردهای قبلی یکسان می‌کند.
func geographyMatchKey(value string) string {
	return strings.ReplaceAll(NormalizeSearch(value), " ", "")
}

func ValidateGeographySeed(seed GeographySeed) error {
	source := seed.Source
	if source.Repository != "https://github.com/sajaddp/list-of-cities-in-Iran" ||
		!geographySourceCommit.MatchString(source.Commit) || source.CityDataset != "cities-filtered" ||
		source.OfficialYear < 1400 || source.OfficialYear > 1500 || len(source.Version) > 32 || source.Version == "" ||
		seed.DatasetID != fmt.Sprintf("iran-geography-%d-v%s", source.OfficialYear, source.Version) ||
		source.ProvinceCount < 1 || source.ProvinceCount > 100 || source.ProvinceCount != len(seed.Provinces) ||
		source.CityCount < 1 || source.CityCount > 10000 {
		return fmt.Errorf("invalid geography source metadata or counts")
	}
	if _, err := parseGeography(nil, seed.Provinces); err != nil {
		return fmt.Errorf("invalid geography directory: %w", err)
	}
	provinceNames := map[string]bool{}
	count := 0
	for _, p := range seed.Provinces {
		key := geographyMatchKey(p.Name)
		if !p.Active || !geographyProvinceID.MatchString(p.ID) || provinceNames[key] || len(p.Cities) == 0 {
			return fmt.Errorf("invalid geography province %s", p.ID)
		}
		provinceNames[key] = true
		cityNames := map[string]bool{}
		for _, c := range p.Cities {
			key := geographyMatchKey(c.Name)
			if !c.Active || !geographyCityID.MatchString(c.ID) || cityNames[key] {
				return fmt.Errorf("invalid or ambiguous geography city %s", c.ID)
			}
			cityNames[key] = true
			count++
		}
	}
	if count != source.CityCount {
		return fmt.Errorf("geography city count mismatch")
	}
	return nil
}

// ادغام افزایشی است: شناسه، نام، وضعیت و مقصدهای قدیمی حذف یا بازنویسی نمی‌شوند.
func MergeGeographySeed(previous []commerce.Province, seed GeographySeed) ([]commerce.Province, GeographyImportResult, error) {
	result := GeographyImportResult{DatasetID: seed.DatasetID}
	if err := ValidateGeographySeed(seed); err != nil {
		return nil, result, err
	}
	out := append([]commerce.Province{}, previous...)
	provinceNames := map[string]int{}
	provinceIDs := map[string]int{}
	allIDs := map[string]bool{}
	cityParents := map[string]string{}
	for i, p := range out {
		out[i].Cities = append([]commerce.City{}, p.Cities...)
		key := geographyMatchKey(p.Name)
		if _, exists := provinceNames[key]; exists {
			return nil, result, fmt.Errorf("ambiguous existing province: %s", p.ID)
		}
		provinceNames[key], provinceIDs[p.ID], allIDs[p.ID] = i, i, true
		for _, c := range p.Cities {
			allIDs[c.ID], cityParents[c.ID] = true, p.ID
		}
	}
	for _, incoming := range seed.Provinces {
		key := geographyMatchKey(incoming.Name)
		index, matched := provinceNames[key]
		if ownIndex, found := provinceIDs[incoming.ID]; found && (!matched || ownIndex != index) {
			return nil, result, fmt.Errorf("province identifier collision: %s", incoming.ID)
		}
		if matched {
			result.ProvincesMatched++
		} else {
			if allIDs[incoming.ID] {
				return nil, result, fmt.Errorf("geography identifier collision: %s", incoming.ID)
			}
			index = len(out)
			out = append(out, commerce.Province{ID: incoming.ID, Name: incoming.Name, Active: true, Cities: []commerce.City{}})
			provinceNames[key], provinceIDs[incoming.ID], allIDs[incoming.ID] = index, index, true
			result.ProvincesAdded++
		}
		p := &out[index]
		cityNames := map[string]int{}
		for j, c := range p.Cities {
			key := geographyMatchKey(c.Name)
			if _, exists := cityNames[key]; exists {
				return nil, result, fmt.Errorf("ambiguous existing city in province %s", p.ID)
			}
			cityNames[key] = j
		}
		for _, incomingCity := range incoming.Cities {
			key := geographyMatchKey(incomingCity.Name)
			cityIndex, matched := cityNames[key]
			if parent, exists := cityParents[incomingCity.ID]; exists && (parent != p.ID || !matched || p.Cities[cityIndex].ID != incomingCity.ID) {
				return nil, result, fmt.Errorf("city identifier collision or changed province: %s", incomingCity.ID)
			}
			if matched {
				result.CitiesMatched++
				continue
			}
			if allIDs[incomingCity.ID] {
				return nil, result, fmt.Errorf("geography identifier collision: %s", incomingCity.ID)
			}
			cityNames[key] = len(p.Cities)
			p.Cities = append(p.Cities, incomingCity)
			allIDs[incomingCity.ID], cityParents[incomingCity.ID] = true, p.ID
			result.CitiesAdded++
		}
	}
	validated, err := parseGeography(previous, out)
	if err != nil {
		return nil, result, fmt.Errorf("merged geography rejected: %w", err)
	}
	result.TotalProvinces = len(validated)
	for _, p := range validated {
		result.TotalCities += len(p.Cities)
	}
	result.AlreadyApplied = result.ProvincesAdded == 0 && result.CitiesAdded == 0
	return validated, result, nil
}

var errGeographyUnchanged = errors.New("geography seed already present")

// این مسیر فقط ابزار نگهداری محلی است؛ endpoint عمومی یا دسترسی فروشنده ایجاد نمی‌کند.
func ApplyGeographySeed(ctx context.Context, repo interface {
	Update(context.Context, func(*Workspace) error) error
}, seed GeographySeed, expectedVersion int, now time.Time) (GeographyImportResult, error) {
	result := GeographyImportResult{DatasetID: seed.DatasetID}
	if err := ValidateGeographySeed(seed); err != nil {
		return result, err
	}
	raw, _ := json.Marshal(seed)
	sum := sha256.Sum256(raw)
	fingerprint := hex.EncodeToString(sum[:])
	err := repo.Update(ctx, func(w *Workspace) error {
		if w.Commerce == nil || w.Commerce.Settings.Version != expectedVersion {
			return ErrConflict
		}
		merged, plan, err := MergeGeographySeed(w.Commerce.Settings.Geography, seed)
		result = plan
		result.SettingsVersion = w.Commerce.Settings.Version
		if err != nil {
			return err
		}
		if result.AlreadyApplied {
			return errGeographyUnchanged
		}
		w.Commerce.Settings.Geography = merged
		w.Commerce.Settings.Version++
		result.SettingsVersion = w.Commerce.Settings.Version
		w.Audit = append(w.Audit, Audit{ID: "iran-geography-" + fingerprint[:24], Actor: "geography-import-cli", Action: "commerce.geography.import", Target: seed.DatasetID, Reason: "merge pinned Iran province/city directory; preserve existing destinations", Fingerprint: fingerprint, At: now})
		return nil
	})
	if errors.Is(err, errGeographyUnchanged) {
		err = nil
	}
	return result, err
}
