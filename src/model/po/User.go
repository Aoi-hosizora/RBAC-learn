package po

type User struct {
	ID       uint32 `gorm:"primary_key;auto_increment"`
	Name     string `gorm:"not_null"`
	Password string `gorm:"not_null"`
	Role     string `gorm:"not_null;default:'normal'"`
}
