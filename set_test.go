package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestContains(t *testing.T) {
	set := NewSet()

	// Oracle testing
	dummy := 0
	in := set.Contains(dummy)
	if in {
		t.Errorf("oracle > set.Contains(%d) = true before insertion", dummy)
	}

	set.Insert(dummy)
	in = set.Contains(dummy)
	if !in {
		t.Errorf("oracle > set.Contains(%d) = false after insertion", dummy)
	}

	// Property testing
	rand.Seed(time.Now().UnixNano())
	const reps = 1000
	for i := 0; i < reps; i++ {
		lastInsert := rand.Int()
		set.Insert(lastInsert)

		in = set.Contains(lastInsert)

		if !in {
			t.Errorf("property > set.Contains(%d) = false after insertion", dummy)
		}
	}
}

func TestInsert(t *testing.T) {
	set := NewSet()

	// Oracle testing
	dummy := 0
	innovative := set.Insert(dummy)

	if !innovative {
		t.Errorf("oracle > set.Insert(%d) = false but was innovative", dummy)
	}

	in := set.Contains(dummy)
	if !in {
		t.Errorf("oracle > set.Contains(%d) = false after insertion", dummy)
	}

	// Duplicate insertion should return false
	innovative = set.Insert(dummy)
	if innovative {
		t.Errorf("oracle > set.Insert(%d) = true but was not innovative", dummy)
	}

	// Property testing
	rand.Seed(time.Now().UnixNano())
	const reps = 1000
	for i := 0; i < reps; i++ {
		val := rand.Int()

		inBefore := set.Contains(val)
		innovative = set.Insert(val)

		if inBefore && innovative {
			t.Errorf("property > included value reported as innovative")
		}
	}
}
