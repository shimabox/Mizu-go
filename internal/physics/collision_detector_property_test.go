package physics

import (
	"slices"
	"strconv"
	"testing"

	"github.com/shimabox/Mizu-go/internal/core"
)

// TestGridCollisionDetector_MatchesBruteForce_Property は
// tests/physics/CollisionDetectorProperty.test.ts を移植したもの:
// いくつかのシードと密に散らばった粒子群に対して、GridCollisionDetector
// は O(n^2) の BruteForceCollisionDetector とまったく同じペアの集合を
// 見つけなければならない。
func TestGridCollisionDetector_MatchesBruteForce_Property(t *testing.T) {
	const width, height = 800.0, 600.0
	seeds := []uint64{1, 2, 3, 4, 5, 6, 7}

	for _, seed := range seeds {
		t.Run(seedTestName(seed), func(t *testing.T) {
			random := core.NewSeededRandom(seed)
			particles := createRandomParticles(500, random, width, height)

			expected := normalizePairs(particles, NewBruteForceCollisionDetector().FindHitPairs(particles))

			grid := NewSpatialGrid(width, height, DefaultCellSize)
			actual := normalizePairs(particles, NewGridCollisionDetector(grid).FindHitPairs(particles))

			if !slices.Equal(actual, expected) {
				t.Fatalf("grid detector disagrees with brute force for seed=%d\n got: %v\nwant: %v", seed, actual, expected)
			}
			// このテストが実際に意味を持つこと(衝突が発生するだけの密度が
			// あること)を確認する。
			if len(expected) == 0 {
				t.Fatalf("seed=%d produced no collisions; test is not exercising anything", seed)
			}
		})
	}
}

func TestGridCollisionDetector_MatchesBruteForce_SparseParticles(t *testing.T) {
	const width, height = 800.0 * 3, 600.0 * 3

	random := core.NewSeededRandom(99)
	particles := createRandomParticles(50, random, width, height)

	expected := normalizePairs(particles, NewBruteForceCollisionDetector().FindHitPairs(particles))

	grid := NewSpatialGrid(width, height, DefaultCellSize)
	actual := normalizePairs(particles, NewGridCollisionDetector(grid).FindHitPairs(particles))

	if !slices.Equal(actual, expected) {
		t.Fatalf("grid detector disagrees with brute force for sparse case\n got: %v\nwant: %v", actual, expected)
	}
}

func seedTestName(seed uint64) string {
	return "seed_" + strconv.FormatUint(seed, 10)
}
