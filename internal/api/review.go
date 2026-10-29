package api

type PurchaseReviewResponse struct {
	Data PurchaseReview `json:"data"`
}

type PurchaseReview struct {
	Summary   string   `json:"summary"`
	NextSteps []string `json:"next_steps"`
	Demo      bool     `json:"demo"`
}
