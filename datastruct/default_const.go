package datastruct

type contextKey string
var (
	//ContextTransactionID for context transaction ID
	ContextTransactionID = contextKey("IMSTransactionID")
)


// Constant Value
const (
	WebTokenKey        string = "Alkjasf2q3BJNDFL@WE14y8dsalj"
	JWTTokenExpiry     int    = 3600 //seconds
	MAXJWTTokenLife    int    = 1    //hours
	MAXJWTTokenCounter int    = 5    //times
	APPID				string = "NEWPPOB"
	SSDBCheckNumberKey  string = "CHKNBR-"
	WARegistered    		string = "1"
	WANotRegistered 		string = "2"
)

