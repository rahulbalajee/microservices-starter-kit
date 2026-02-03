package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string //ex: van, sedan, luxury
	TotalPriceInCents float64
	Expires           time.Time
}
