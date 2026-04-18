package license

// Bundle constants matching the pricing tiers.
const (
	BundleFree  = "free"             // OSS tools, no license needed
	BundleCore  = "core"             // ee + performance + audit + partitioning
	BundleHA    = "ha"               // dataguard + backup + rac + clusterware + asm
	BundleOps   = "ops"              // provision + patch + migration + datapump + goldengate + oem
	BundlePGPro = "pg_professional"  // mcp-postgres-enterprise
)

// DomainToBundle maps tool domains to their required bundle.
// OSS domains (db-read, linux, monitor-agent) require no license.
var DomainToBundle = map[string]string{
	// Free (OSS)
	"db-read":         BundleFree,
	"linux":           BundleFree,
	"monitor-agent":   BundleFree,

	// Core bundle
	"db-mutate":       BundleCore,
	"performance":     BundleCore,
	"audit":           BundleCore,
	"partitioning":    BundleCore,
	"monitor-central": BundleCore,

	// HA bundle
	"dataguard":       BundleHA,
	"backup":          BundleHA,
	"rac":             BundleHA,
	"clusterware":     BundleHA,
	"asm":             BundleHA,

	// Ops bundle
	"provision":       BundleOps,
	"patch":           BundleOps,
	"migration":       BundleOps,
	"datapump":        BundleOps,
	"goldengate":      BundleOps,
	"oem":             BundleOps,

	// PG Professional
	"pg-enterprise":   BundlePGPro,
}

// RequiredBundle returns the bundle required for a given tool domain.
// Returns BundleFree if the domain is not mapped (safe fallback).
func RequiredBundle(domain string) string {
	if b, ok := DomainToBundle[domain]; ok {
		return b
	}
	return BundleFree
}
