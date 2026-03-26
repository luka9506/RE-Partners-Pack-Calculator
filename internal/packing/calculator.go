package packing

import (
	"errors"
	"fmt"
	"sort"
)

type Result struct {
	OrderedQuantity int           `json:"ordered_quantity"`
	TotalItems      int           `json:"total_items"`
	Overfill        int           `json:"overfill"`
	Packs           []PackSummary `json:"packs"`
}

type PackSummary struct {
	PackSize int `json:"pack_size"`
	Count    int `json:"count"`
}

type Calculator struct {
	packSizes []int
}

type candidate struct {
	totalItems int
	packCount  int
	counts     []int
}

func NewCalculator(packSizes []int) (*Calculator, error) {
	if len(packSizes) == 0 {
		return nil, errors.New("at least one pack size is required")
	}

	// Normalize once so the search can treat pack sizes as an ordered set.
	normalized := make([]int, len(packSizes))
	copy(normalized, packSizes)
	sort.Ints(normalized)

	for index, size := range normalized {
		if size <= 0 {
			return nil, fmt.Errorf("pack sizes must be positive: %d", size)
		}
		if index > 0 && normalized[index-1] == size {
			return nil, fmt.Errorf("duplicate pack size: %d", size)
		}
	}

	return &Calculator{packSizes: normalized}, nil
}

func (c *Calculator) Calculate(orderQuantity int) (Result, error) {
	if orderQuantity <= 0 {
		return Result{}, errors.New("order quantity must be greater than zero")
	}

	// Any optimal answer must land before orderQuantity+maxPack because one extra
	// largest pack is always enough to cross the target from some smaller total.
	maxPack := c.packSizes[len(c.packSizes)-1]
	upperBound := orderQuantity + maxPack - 1
	bestByTotal := make([]*candidate, upperBound+1)
	bestByTotal[0] = &candidate{counts: make([]int, len(c.packSizes))}

	// For each reachable total, keep only the solution with the fewest packs.
	// Scanning totals upward afterwards guarantees that the first reachable total
	// at or above the order quantity also minimizes shipped items.
	for total := 0; total <= upperBound; total++ {
		current := bestByTotal[total]
		if current == nil {
			continue
		}

		for index, size := range c.packSizes {
			nextTotal := total + size
			if nextTotal > upperBound {
				continue
			}

			nextCounts := make([]int, len(current.counts))
			copy(nextCounts, current.counts)
			nextCounts[index]++

			nextCandidate := &candidate{
				totalItems: nextTotal,
				packCount:  current.packCount + 1,
				counts:     nextCounts,
			}

			existing := bestByTotal[nextTotal]
			if existing == nil || nextCandidate.packCount < existing.packCount {
				bestByTotal[nextTotal] = nextCandidate
			}
		}
	}

	for total := orderQuantity; total <= upperBound; total++ {
		best := bestByTotal[total]
		if best == nil {
			continue
		}
		return c.buildResult(orderQuantity, best), nil
	}

	return Result{}, errors.New("no valid pack combination found")
}

func (c *Calculator) buildResult(orderQuantity int, choice *candidate) Result {
	// Return packs largest-first because that is the clearest presentation for
	// both the UI and API consumers.
	packs := make([]PackSummary, 0, len(c.packSizes))
	for index := len(c.packSizes) - 1; index >= 0; index-- {
		count := choice.counts[index]
		if count == 0 {
			continue
		}
		packs = append(packs, PackSummary{
			PackSize: c.packSizes[index],
			Count:    count,
		})
	}

	return Result{
		OrderedQuantity: orderQuantity,
		TotalItems:      choice.totalItems,
		Overfill:        choice.totalItems - orderQuantity,
		Packs:           packs,
	}
}
