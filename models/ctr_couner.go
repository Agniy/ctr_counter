package models

import "time"

type CtrCounter struct{
	Id uint32
	ProductId uint32
	Date time.Time
	Show uint32
	Click uint32
	CatalogClick uint32
	Ctr uint32
	Change bool
}


