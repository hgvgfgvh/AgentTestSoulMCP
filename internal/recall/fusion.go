package recall

// fuseRRF 多路排序 Reciprocal Rank Fusion（index -> 事实下标）。
func fuseRRF(rankings [][]int, topN, k int) []int {
	if k <= 0 {
		k = 60
	}
	scores := map[int]float64{}
	for _, ranks := range rankings {
		for rank, idx := range ranks {
			scores[idx] += 1.0 / (float64(k) + float64(rank+1))
		}
	}
	if len(scores) == 0 {
		return nil
	}
	type pair struct {
		idx   int
		score float64
	}
	pairs := make([]pair, 0, len(scores))
	for idx, sc := range scores {
		pairs = append(pairs, pair{idx, sc})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].score > pairs[i].score {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	if topN <= 0 || len(pairs) <= topN {
		out := make([]int, len(pairs))
		for i, p := range pairs {
			out[i] = p.idx
		}
		return out
	}
	out := make([]int, topN)
	for i := 0; i < topN; i++ {
		out[i] = pairs[i].idx
	}
	return out
}
