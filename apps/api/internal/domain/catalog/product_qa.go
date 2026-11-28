package catalog

import "time"

type QAStatus string

const (
	QAPending  QAStatus = "pending"
	QAApproved QAStatus = "approved"
	QARejected QAStatus = "rejected"
)

const ProductQATextLimit = 2000

type ProductQAContent struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Status    QAStatus  `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductQuestion struct {
	ProductQAContent
	Answers []ProductQAContent `json:"answers"`
	Version int                `json:"version"`
	Cursor  int64              `json:"-"`
}
