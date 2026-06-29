package au_image

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
)

// xmlOneoffInventory mirrors the on-disk shape of
// `<ORACLE_HOME>/inventory/oneoffs/<id>/etc/config/inventory.xml`.
type xmlOneoffInventory struct {
	XMLName          xml.Name      `xml:"oneoff_inventory"`
	PatchDescription string        `xml:"patch_description"`
	PatchID          xmlPatchID    `xml:"patch_id"`
	UniquePatchID    string        `xml:"unique_patch_id"`
	MinOpatchVersion string        `xml:"minimum_opatch_version"`
	DateOfPatch      xmlDate       `xml:"date_of_patch"`
	BaseBugs         xmlBaseBugs   `xml:"base_bugs"`
}

type xmlPatchID struct {
	Number string `xml:"number,attr"`
}

type xmlDate struct {
	Day   string `xml:"day,attr"`
	Month string `xml:"month,attr"`
	Time  string `xml:"time,attr"`
	Year  string `xml:"year,attr"`
	Zone  string `xml:"zone,attr"`
}

type xmlBaseBugs struct {
	Bugs []xmlBug `xml:"bug"`
}

type xmlBug struct {
	Number      string `xml:"number,attr"`
	Description string `xml:"description,attr"`
}

// dbRUDescriptionRE matches the canonical "Database Release Update :
// 19.<minor>.0.0.<date> (<patch_id>)" pattern. Capture[1] = minor, capture[2] = patch_id.
var dbRUDescriptionRE = regexp.MustCompile(`^Database Release Update : 19\.(\d+)\.\d+\.\d+\.\d+ \((\d+)\)$`)

// ocwBundledRE matches OCW Release Updates which ship pre-bundled with
// the base zip (e.g. "OCW RELEASE UPDATE 19.3.0.0.0").
var ocwBundledRE = regexp.MustCompile(`^OCW RELEASE UPDATE\b`)

// ParseOneoffInventory parses one inventory.xml from the on-disk
// oneoffs/<id>/etc/config/ tree and returns the per-patch record.
func ParseOneoffInventory(r io.Reader) (*OneoffPatch, error) {
	var raw xmlOneoffInventory
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse oneoff inventory: %w", err)
	}
	if strings.TrimSpace(raw.PatchID.Number) == "" {
		return nil, fmt.Errorf("parse oneoff inventory: missing <patch_id number=...>")
	}
	op := &OneoffPatch{
		PatchID:          strings.TrimSpace(raw.PatchID.Number),
		UniquePatchID:    strings.TrimSpace(raw.UniquePatchID),
		Description:      strings.TrimSpace(raw.PatchDescription),
		MinOpatchVersion: strings.TrimSpace(raw.MinOpatchVersion),
	}
	if t, err := parseOraclePatchDate(raw.DateOfPatch); err == nil {
		op.AppliedDate = t
	}
	if len(raw.BaseBugs.Bugs) > 0 {
		op.BaseBugs = make([]BaseBug, 0, len(raw.BaseBugs.Bugs))
		for _, b := range raw.BaseBugs.Bugs {
			op.BaseBugs = append(op.BaseBugs, BaseBug{
				Number:      b.Number,
				Description: b.Description,
			})
		}
	}
	return op, nil
}

// parseOraclePatchDate parses the day/month/year triplet from the
// `<date_of_patch>` element. Oracle uses 3-letter English month
// abbreviations and an optional named timezone we map heuristically.
func parseOraclePatchDate(d xmlDate) (time.Time, error) {
	if d.Day == "" || d.Month == "" || d.Year == "" {
		return time.Time{}, fmt.Errorf("incomplete date")
	}
	// Build a parseable string.
	tm := strings.TrimSpace(strings.TrimSuffix(d.Time, " hrs"))
	if tm == "" {
		tm = "00:00:00"
	}
	zone := strings.TrimSpace(d.Zone)
	if zone == "" {
		zone = "UTC"
	}
	// Use a named layout reference. Oracle zone strings ("UTC",
	// "PST8PDT") are not all parseable by Go's MST layout, so try a
	// permissive ladder.
	candidates := []string{
		fmt.Sprintf("%s %s %s %s %s", d.Day, d.Month, d.Year, tm, zone),
		fmt.Sprintf("%s %s %s %s UTC", d.Day, d.Month, d.Year, tm),
	}
	layouts := []string{
		"2 Jan 2006 15:04:05 MST",
		"2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05",
	}
	for _, s := range candidates {
		for _, l := range layouts {
			if t, err := time.Parse(l, s); err == nil {
				return t, nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date: %+v", d)
}

// ParseCompsXML extracts the PATCH_LEVEL attribute from the first
// <COMP NAME="oracle.server"> element. Returns "" (no error) when the
// attribute is absent — older base homes do not stamp PATCH_LEVEL until
// at least one OPatch apply has run.
//
// Implementation note: comps.xml is huge (50+MB on a real home) and we
// only care about the first oracle.server <COMP> attributes, so we
// scan tokens and return as soon as we find it. This also makes the
// parser robust to truncation (callers may stream just the head).
func ParseCompsXML(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			return "", fmt.Errorf("parse comps.xml: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "COMP" {
			continue
		}
		var name, level string
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "NAME":
				name = a.Value
			case "PATCH_LEVEL":
				level = a.Value
			}
		}
		if name == "oracle.server" {
			return strings.TrimSpace(level), nil
		}
	}
}

// MapPatchLevelToRU translates the comps.xml PATCH_LEVEL stamp to an
// RU label. The mapping is empirical (Oracle does not document it):
//
//	"0" → "19.3 (base)"
//	"1" → "19.26"
//	"2" → "19.27"
//	"3" → "19.28"
//
// Used as a fallback when no <oneoff> matches the DB RU regex.
func MapPatchLevelToRU(level string) (string, error) {
	switch strings.TrimSpace(level) {
	case "0":
		return "19.3 (base)", nil
	case "1":
		return "19.26", nil
	case "2":
		return "19.27", nil
	case "3":
		return "19.28", nil
	case "":
		return "", fmt.Errorf("comps.xml has no PATCH_LEVEL stamp (base home, no RU applied)")
	default:
		return "", fmt.Errorf("unknown PATCH_LEVEL %q (extend MapPatchLevelToRU when new RUs ship)", level)
	}
}

// ClassifyPatches takes the union of patches discovered under
// inventory/oneoffs/* and partitions them into:
//
//   - the active RU (the highest 19.<minor> Database Release Update),
//   - real one-offs (operator-chosen patches),
//
// patches matching the OPatch-bundled patterns (older RU shadow
// 19.3.x.x or OCW Release Update) are flagged via IsOpatchBundled
// and excluded from both the active RU choice and the returned oneoffs.
//
// Returns an error if multiple distinct candidate active RUs are found
// at the same minor (e.g. two "Database Release Update : 19.28..." on
// disk — should be impossible).
func ClassifyPatches(all []OneoffPatch) (*RUPatch, []OneoffPatch, error) {
	type ruCandidate struct {
		minor       int
		patchID     string
		description string
	}
	var ruCands []ruCandidate
	var oneoffs []OneoffPatch

	for _, p := range all {
		desc := p.Description

		// OCW base bundles: ship inside the base 19.3 zip's
		// inventory/oneoffs already populated. Not a real one-off.
		if ocwBundledRE.MatchString(desc) {
			p.IsOpatchBundled = true
			oneoffs = append(oneoffs, p)
			continue
		}

		if m := dbRUDescriptionRE.FindStringSubmatch(desc); m != nil {
			minor := atoiSafe(m[1])
			// 19.3.x.x is the base-bundled "shadow" RU; the real
			// RU is always strictly higher.
			if minor <= 3 {
				p.IsOpatchBundled = true
				oneoffs = append(oneoffs, p)
				continue
			}
			ruCands = append(ruCands, ruCandidate{
				minor:       minor,
				patchID:     p.PatchID,
				description: p.Description,
			})
			continue
		}

		// Everything else is an operator-chosen one-off.
		oneoffs = append(oneoffs, p)
	}

	if len(ruCands) == 0 {
		return nil, nil, fmt.Errorf("classify: no active RU candidate found among %d patches (expected one matching %q)", len(all), dbRUDescriptionRE)
	}
	// Sort descending by minor — pick the highest as the active RU.
	sort.Slice(ruCands, func(i, j int) bool { return ruCands[i].minor > ruCands[j].minor })
	if len(ruCands) > 1 && ruCands[0].minor == ruCands[1].minor {
		return nil, nil, fmt.Errorf("classify: multiple RU candidates at minor 19.%d: %s, %s", ruCands[0].minor, ruCands[0].patchID, ruCands[1].patchID)
	}
	chosen := ruCands[0]
	ru := &RUPatch{
		Label:       fmt.Sprintf("19.%d", chosen.minor),
		PatchID:     chosen.patchID,
		Description: chosen.description,
	}
	// Any other RU candidates (lower minor) get folded into the
	// bundled list for transparency.
	for _, rc := range ruCands[1:] {
		oneoffs = append(oneoffs, OneoffPatch{
			PatchID:         rc.patchID,
			Description:     rc.description,
			IsOpatchBundled: true,
		})
	}
	return ru, oneoffs, nil
}

// ComputeAuID returns the deterministic au-image identifier.
// Format:
//
//	au<RU-label>v<rev>                (no real one-offs)
//	au<RU-label>+<sha[:6]>v<rev>      (with real one-offs)
//
// "Real" means !IsOpatchBundled. PatchIDs are sorted lexically before
// hashing so the ID is order-independent.
func ComputeAuID(ru *RUPatch, oneoffs []OneoffPatch, rev int) string {
	if ru == nil {
		return ""
	}
	var realIDs []string
	for _, o := range oneoffs {
		if o.IsOpatchBundled {
			continue
		}
		realIDs = append(realIDs, o.PatchID)
	}
	if len(realIDs) == 0 {
		return fmt.Sprintf("au%sv%d", ru.Label, rev)
	}
	sort.Strings(realIDs)
	sum := sha256.Sum256([]byte(strings.Join(realIDs, ",")))
	return fmt.Sprintf("au%s+%sv%d", ru.Label, hex.EncodeToString(sum[:])[:6], rev)
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
