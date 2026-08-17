package history

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestStoreAddAndRetrieve(t *testing.T) {

	store := NewStore(3)

	store.Add(domain.HealthResult{
		HealthScore: 100,
	})

	store.Add(domain.HealthResult{
		HealthScore: 90,
	})

	results := store.GetAll()

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}
}

func TestStoreKeepsLimit(t *testing.T) {

	store := NewStore(2)

	store.Add(domain.HealthResult{
		HealthScore: 100,
	})

	store.Add(domain.HealthResult{
		HealthScore: 90,
	})

	store.Add(domain.HealthResult{
		HealthScore: 80,
	})

	results := store.GetAll()

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}

	if results[0].HealthScore != 90 {
		t.Errorf(
			"unexpected oldest value: %d",
			results[0].HealthScore,
		)
	}
}

func TestStoreLatest(t *testing.T) {

	store := NewStore(5)

	store.Add(domain.HealthResult{
		HealthScore: 75,
	})

	result, ok := store.Latest()

	if !ok {
		t.Fatal("expected result")
	}

	if result.HealthScore != 75 {
		t.Errorf(
			"score mismatch: got %d",
			result.HealthScore,
		)
	}
}
