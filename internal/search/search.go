package search

type SearchDoc struct {
	ID       string   `json:"id"`
	SourceID string   `json:"source_id"`
	PartIdxs []int32  `json:"part_idxs"`
	Topic    string   `json:"topic"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
	Score    float32  `json:"score"`
}
