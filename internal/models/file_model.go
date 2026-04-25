package models

type Upload_Session struct {
	Id           string   `json:"uuid"`
	User_Id      string   `json:"uuid"`
	Status       string `json:"status"`
	Total_Amount int    `json:"total_amount"`
	Total_Sheets int    `json:"total_sheets"`
	Token        string `json:"token"`
}
