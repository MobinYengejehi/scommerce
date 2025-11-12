package scommerce

import (
	"math"
	"strings"
)

type QueueOrder string

const (
	QueueOrderAscending  QueueOrder = "asc"
	QueueOrderDescending QueueOrder = "desc"
)

func (q QueueOrder) String() string {
	qs := strings.ToLower(string(q))
	if qs == "asc" {
		return "asc"
	} else if qs == "desc" {
		return "desc"
	}

	return "unknown.....__error__"
}

func GetSafeLimit(limit int64) int64 {
	return int64(math.Min(float64(limit), 500))
}
