package bus

type Bus struct {
	ID                 uint   `gorm:"primaryKey;autoincrement"`
	Model              string `gorm:"type:varchar(255); not null"`
	ImageURL           string `gorm:"type:varchar(255); not null"`
	IsActive           bool   `gorm:" not null"`
	RegistrationNumber string `gorm:"type:varchar(8); not null"`
	Seats              []Seat `gorm:"foreignKey:BusID"`
}

type Seat struct {
	ID       uint  `gorm:"primaryKey;autoincrement"`
	BusID    uint  `gorm:"not null"`
	Number   uint8 `gorm:"not null"`
	Window   bool  `gorm:"not null"`
	Reserved bool  `gorm:"not null"`
}
