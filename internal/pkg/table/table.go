package table

import "github.com/cravtos/arithmetic/internal/pkg/config"

// maxTotalCount is the maximum value of the sum of counts
const maxTotalCount uint64 = (1 << config.CountBitsUsed) - 1

// ABCSize is the size of the alphabet
const ABCSize = 256

// Table holding counts of characters
type Table struct {
	count      [ABCSize + 1]uint64
	interval   [ABCSize + 1]uint64
	totalCount uint64
}

// NewTable constructs new encoding table
func NewTable() *Table {
	t := &Table{}
	for i := 1; i <= ABCSize; i++ {
		t.count[i] = 1
		t.interval[i] = t.interval[i-1] + t.count[i]
	}
	t.totalCount = ABCSize

	return t
}

// updateRanges makes symbols ranges valid
func (t *Table) updateRanges(fromSymbol uint8) {
	for i := uint(fromSymbol) + 1; i <= ABCSize; i++ {
		t.interval[i] = t.interval[i-1] + t.count[i]
	}
}

// UpdateCount updates symbol count and ranges
// If Table.totalCount is too big, Table.count are normalized
func (t *Table) UpdateCount(symbol uint8) {
	t.count[symbol+1]++
	t.totalCount++

	if t.totalCount >= maxTotalCount {
		t.totalCount = 0

		for i := 1; i <= ABCSize; i++ {
			t.count[i] /= config.CountDenominator

			if t.count[i] == 0 {
				t.count[i] = 1
			}

			t.totalCount += t.count[i]
		}

		t.updateRanges(0)
		return
	}

	t.updateRanges(symbol)
}

// GetInterval returns interval end for given symbol
func (t *Table) GetInterval(symbol uint8) uint64 {
	return t.interval[uint(symbol)+1]
}

// GetSymbol returns symbol with corresponding interval
func (t *Table) GetSymbol(interval uint64) uint8 {
	symbol := 1
	for t.interval[symbol] <= interval {
		symbol++
	}
	return uint8(symbol - 1)
}
