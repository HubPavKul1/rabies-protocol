package domain

// import "time"

type Veterinarian struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Position string `json:"position"`
}

type ClinicConfig struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}