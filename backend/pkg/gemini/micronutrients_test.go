package gemini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMicronutrients_BothNil(t *testing.T) {
	result := MergeMicronutrients(nil, nil)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestMergeMicronutrients_DstNilSrcNonNil(t *testing.T) {
	src := map[string]float64{"iron_mg": 3.0}
	result := MergeMicronutrients(nil, src)
	assert.Equal(t, 3.0, result["iron_mg"])
}

func TestMergeMicronutrients_DstNonNilSrcNil(t *testing.T) {
	dst := map[string]float64{"iron_mg": 5.0}
	result := MergeMicronutrients(dst, nil)
	assert.Equal(t, 5.0, result["iron_mg"])
}

func TestMergeMicronutrients_OverlappingKeys(t *testing.T) {
	dst := map[string]float64{"iron_mg": 2.0, "calcium_mg": 100.0}
	src := map[string]float64{"iron_mg": 3.0, "calcium_mg": 200.0}
	result := MergeMicronutrients(dst, src)
	assert.Equal(t, 5.0, result["iron_mg"])
	assert.Equal(t, 300.0, result["calcium_mg"])
}

func TestMergeMicronutrients_NonOverlappingKeys(t *testing.T) {
	dst := map[string]float64{"iron_mg": 2.0}
	src := map[string]float64{"calcium_mg": 100.0}
	result := MergeMicronutrients(dst, src)
	assert.Equal(t, 2.0, result["iron_mg"])
	assert.Equal(t, 100.0, result["calcium_mg"])
}

func TestMergeMicronutrients_Immutability(t *testing.T) {
	dst := map[string]float64{"iron_mg": 2.0}
	src := map[string]float64{"iron_mg": 3.0}
	originalDst := dst["iron_mg"]
	originalSrc := src["iron_mg"]

	MergeMicronutrients(dst, src)

	assert.Equal(t, originalDst, dst["iron_mg"], "dst must not be mutated")
	assert.Equal(t, originalSrc, src["iron_mg"], "src must not be mutated")
}

func TestMergeMicronutrients_EmptyMaps(t *testing.T) {
	dst := map[string]float64{}
	src := map[string]float64{}
	result := MergeMicronutrients(dst, src)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestValidMicronutrientKeys(t *testing.T) {
	keys := ValidMicronutrientKeys()
	assert.Len(t, keys, len(AllMicronutrients))
	for _, m := range AllMicronutrients {
		assert.True(t, keys[string(m.Key)])
	}
	assert.False(t, keys["invalid_key"])
}

func TestDefaultMicronutrientTargets(t *testing.T) {
	targets := DefaultMicronutrientTargets()
	assert.Len(t, targets, len(AllMicronutrients))
	for _, m := range AllMicronutrients {
		assert.Equal(t, m.DefaultTarget, targets[string(m.Key)])
	}
}
