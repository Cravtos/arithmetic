package config

// IntervalBitsUsed determines how many bits is used by interval
// Should be greater than CountBitsUsed, less than 64
// Also CountBitsUsed + IntervalBitsUsed should be < 64
// Otherwise there will be overflows
const IntervalBitsUsed = 32

// CountDenominator is number on which count will be divided for normalization
const CountDenominator = 2

// CountBitsUsed determines bits for maximal sum of all symbol counts
const CountBitsUsed = 16

// UpdateRangesRate determines how often ranges are recalculated
const UpdateRangesRate = 1000
