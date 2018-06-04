package main

//=======================================
// Chain Structure Definitions
//=======================================
type ChainCode struct {
}

//=======================================
//	Product Structure Declaration
//=======================================
type Product struct {
	ManufacturerName    string				`json:"manufacturer"`
	ManufacturerID		string				`json:"manufacturer_id"`
	Barcode				string				`json:"bar_code"`
	ProductName  		string 				`json:"product_name"`
	ProductID 			string 				`json:"product_id"`
	Code  				string 				`json:"code"`
	Type  				string 				`json:"type"`
	NetWeight      		float32    			`json:"net_weight"`
	Description         string				`json:"description"`
	Price          		float32    			`json:"price"`
	MfgDate				uint64				`json:"mfg_date"`
	ExpiryDate			uint64				`json:"expiry_date"`
}

//=======================================
//	Cart Structure Declaration
//=======================================
type Cart struct {
	Products			map[string]Product	`json:"products"`
	Status				string				`json:"status"`
	TotalPrice			float32				`json:"total_price"`
	TransactionID		string				`json:"transaction_id"`
}

//=======================================
//	User Structure Declaration
//=======================================
type User struct {
	UserID				string				`json:"user_id"`
	CartDetail			[]Cart				`json:"cart_detail"`
} 

