package license

import "fmt"

// EntityLimit defines the included and max limits for an entity type.
type EntityLimit struct {
	Included int `json:"included"`
	Max      int `json:"max"`
}

// EntityClaims maps entity type names to their limits.
// Types: oracle_database, pg_database, host, exadata, oda, zdlra.
type EntityClaims map[string]EntityLimit

// GraceClaim defines the overage grace for entity limits.
type GraceClaim struct {
	OverageEntities int `json:"overage_entities"`
	OverageDays     int `json:"overage_days"`
}

// CheckEntityLimit verifies that the current count for an entity type
// is within the license limit (max + overage grace).
// Returns nil if allowed, error if rejected.
func CheckEntityLimit(entities EntityClaims, grace GraceClaim, entityType string, currentCount int) error {
	limit, ok := entities[entityType]
	if !ok {
		return fmt.Errorf("entity type %q not licensed (available types: %v)", entityType, entityTypeKeys(entities))
	}

	hardMax := limit.Max + grace.OverageEntities
	if currentCount <= limit.Max {
		return nil // within normal limit
	}
	if currentCount <= hardMax {
		// Within overage grace — allowed but should warn
		return nil
	}
	return fmt.Errorf("entity limit exceeded for %s: %d registered, max %d + %d overage = %d (upgrade license for more)",
		entityType, currentCount, limit.Max, grace.OverageEntities, hardMax)
}

// EntityCapacity returns remaining capacity per entity type.
func EntityCapacity(entities EntityClaims, registered map[string]int) map[string]int {
	capacity := make(map[string]int, len(entities))
	for entityType, limit := range entities {
		used := registered[entityType]
		remaining := limit.Max - used
		if remaining < 0 {
			remaining = 0
		}
		capacity[entityType] = remaining
	}
	return capacity
}

// CheckBundleEntitlement verifies that the license includes the specified bundle.
func CheckBundleEntitlement(licensedBundles []string, required string) error {
	for _, b := range licensedBundles {
		if b == required {
			return nil
		}
	}
	return fmt.Errorf("bundle %q not licensed (have: %v)", required, licensedBundles)
}

// RegistrationResult is the response from a target registration check.
type RegistrationResult struct {
	Allowed          bool   `json:"allowed"`
	EntityType       string `json:"entity_type"`
	Remaining        int    `json:"remaining"`         // capacity left after this registration
	InOverage        bool   `json:"in_overage"`        // true if in overage grace
	OverageRemaining int    `json:"overage_remaining"` // overage slots left
}

// CheckRegistration verifies whether a new target of the given entity type
// can be registered under current license limits.
func CheckRegistration(
	entities EntityClaims,
	grace GraceClaim,
	registered map[string]int,
	entityType string,
) (*RegistrationResult, error) {
	limit, ok := entities[entityType]
	if !ok {
		return nil, fmt.Errorf("entity type %q not licensed", entityType)
	}

	currentCount := registered[entityType]
	newCount := currentCount + 1
	hardMax := limit.Max + grace.OverageEntities

	if newCount > hardMax {
		return nil, fmt.Errorf("entity limit exceeded for %s: %d registered + 1 new > %d max + %d overage",
			entityType, currentCount, limit.Max, grace.OverageEntities)
	}

	result := &RegistrationResult{
		Allowed:    true,
		EntityType: entityType,
	}

	if newCount > limit.Max {
		result.InOverage = true
		result.OverageRemaining = hardMax - newCount
		result.Remaining = 0
	} else {
		result.Remaining = limit.Max - newCount
	}

	return result, nil
}

func entityTypeKeys(entities EntityClaims) []string {
	keys := make([]string, 0, len(entities))
	for k := range entities {
		keys = append(keys, k)
	}
	return keys
}
