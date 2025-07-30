// file: service/utils/diff.go
package utils

// Float32MapsEqual compares two map[string]float32 for equality.
func Float32MapsEqual(a, b map[string]float32) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// MapsEqual recursively compares two map[string]interface{} for equality, including nested maps.
func MapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		switch va := v.(type) {
		case map[string]interface{}:
			bvMap, ok := bv.(map[string]interface{})
			if !ok || !MapsEqual(va, bvMap) {
				return false
			}
		case map[string]float32:
			bvMap, ok := bv.(map[string]float32)
			if !ok || !Float32MapsEqual(va, bvMap) {
				return false
			}
		default:
			if va != bv {
				return false
			}
		}
	}
	return true
}

// collectChangedFields recursively collects changed fields into the changed map.
func CollectChangedFields(curr, prev map[string]interface{}, changed map[string]interface{}, prefix string) {
	for k, v := range curr {
		if pv, ok := prev[k]; ok {
			switch va := v.(type) {
			case map[string]interface{}:
				bv, ok := pv.(map[string]interface{})
				if ok {
					// Only add prefix for nested maps (not float groups)
					CollectChangedFields(va, bv, changed, prefix+k+".")
				}
			case map[string]float32:
				bv, ok := pv.(map[string]float32)
				if ok {
					for fk, fv := range va {
						if bv[fk] != fv {
							// Use fk as the full field name (do not prefix with group)
							changed[fk] = fv
						}
					}
				}
			default:
				fullKey := k
				if prefix != "" {
					fullKey = prefix + k
				}
				if va != pv {
					changed[fullKey] = va
				}
			}
		} else {
			// New key
			switch va := v.(type) {
			case map[string]interface{}:
				CollectChangedFields(va, map[string]interface{}{}, changed, prefix+k+".")
			case map[string]float32:
				for fk, fv := range va {
					changed[fk] = fv
				}
			default:
				fullKey := k
				if prefix != "" {
					fullKey = prefix + k
				}
				changed[fullKey] = va
			}
		}
	}
}
