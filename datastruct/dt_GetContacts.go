package datastruct

type GetContactResponse struct {
	Type    string         `json:"type"`
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Detail  []ListContacts `json:"detail"`
}

type ListContacts struct {
	ID      string
	Name    string
	IsGroup bool
}

type ContactDetailResponse struct {
	Type    string             `json:"type"`
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Detail  ContactInformation `json:"detail"`
}

type ContactInformation struct {
	ID           ContactID `json:"id"`
	Number       string    `json:"number"`
	PushName     string    `json:"pushname"`
	VerifiedName string    `json:"verifiedName"`
	IsBusiness   bool      `json:"isBusiness"`
}

type ContactID struct {
	Server     string `json:"server"`
	User       string `json:"user"`
	Serialized string `json:"_serialized"`
}
