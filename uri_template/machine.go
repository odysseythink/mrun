package uritemplate

// threadList implements https://research.swtch.com/sparse.
type threadList struct {
	dense  []threadEntry
	sparse []uint32
}

type threadEntry struct {
	pc uint32
	t  *thread
}

type thread struct {
	op  *progOp
	cap map[string][]int
}
