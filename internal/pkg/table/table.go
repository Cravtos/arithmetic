package table

// Value borders
const countBitsUsed = 24
const maxTotalCount uint64 = (1 << countBitsUsed) - 1

// Multipliers
const countDenominator = 2

// Size of the alphabet
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
	for i := 1; i < ABCSize+1; i++ {
		t.count[i] = 1
		t.interval[i] = t.interval[i-1] + t.count[i]
	}
	t.totalCount = ABCSize

	return t
}

// updateRanges makes symbols ranges valid
func (t *Table) updateRanges(fromSymbol uint8) {
	for i := uint(fromSymbol) + 1; i < ABCSize+1; i++ {
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

		for i := 1; i < ABCSize+1; i++ {
			t.count[i] /= countDenominator

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
