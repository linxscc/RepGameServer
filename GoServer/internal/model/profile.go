package model

// ProfileInfo 个人信息
type ProfileInfo struct {
	ID        int    `json:"id"`
	FullName  string `json:"fullName"`
	Title     string `json:"title"`
	Tagline   string `json:"tagline"`
	AboutText string `json:"aboutText"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Languages string `json:"languages"`
}
