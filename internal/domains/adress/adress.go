package adress

type Adress struct {
	ID             uint   `gorm:"primaryKey;autoincrement"`
	CustomerID     uint   `gorm:"not null"`
	Country        string `gorm:"type:varchar(56);not null"`
	City           string `gorm:"type:varchar(56);not null"`
	Street         string `gorm:"type:varchar(255);not null"`
	HouseNumber    uint16 `gorm:"not null"`
	Apartment      uint16
	GoogleMapsLink string `gorm:"type:varchar(255);not null"`
	Active         bool   `gorm:"not null"`
}
