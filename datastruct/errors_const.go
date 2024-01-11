package datastruct

// Internal Error message definition
const (
	//Success
	ErrSuccess  int    = 0
	DescSuccess string = "Success"

	ErrFailed  int    = 1
	DescFailed string = "Failed"

	//Service Request Error
	ErrGetMappingValueIncompleteRequest  int    = 1100
	DescGetMappingValueIncompleteRequest string = "Incomplete Request"

	ErrGetMappingDistinctIncompleteRequest  int    = 1101
	DescGetMappingDistinctIncompleteRequest string = "Incomplete Request"

	ErrGetCustStatusIncompleteRequest  int    = 1102
	DescGetCustStatusIncompleteRequest string = "Incomplete Request"

	//Functional Errors
	ErrStringConvert  int    = 3001
	DescStringConvert string = "Convert Error"

	ErrUnknown  int    = 3002
	DescUnknown string = "[3002]Unknown Error"

	ErrData  int    = 3003
	DescData string = "Incomplete Request"

	//Data Error
	ErrNoData  int    = 6990
	DescNoData string = "No Data"

	ErrInvalidData  int    = 6010
	DescInvalidData string = "Invalid Data"

	ErrDataExists  int    = 6011
	DescDataExists string = "Data is already exists"

	ErrDeleteData  int    = 6012
	DescDeleteData string = "Cannot delete the data"

	//Query Error
	ErrFailedQuery  int    = 7010
	DescFailedQuery string = "Read Database Value Error"

	//Redis Error
	ErrFailedRedis  int    = 7020
	DescFailedRedis string = "Redis Error"
	//Unauthorized
	ErrUnauthorized  int    = 8010
	DescUnauthorized string = "Unauthorized"

	ErrWrongUserPassword  int    = 8020
	DescWrongUserPassword string = "Wrong User Password"

	ErrLoginFailed  int    = 8110
	DescLoginFailed string = "Login Failed"

	ErrLoginFailedNoToken  int    = 8120
	DescLoginFailedNoToken string = "Login Failed - No Token"

	ErrFailedGenerateToken  int    = 8121
	DescFailedGenerateToken string = "Error while generating token"

	ErrMaxToken  int    = 8122
	DescMaxToken string = "Token max attemp reached"

	ErrNotAuthorized  int    = 8123
	DescNotAuthorized string = "Not Authorized"

	ErrValidateToken     int    = 8130
	DescErrValidateToken string = "Cannot Validate Token"

	ErrTokenExpired  int    = 8131
	DescTokenExpired string = "Token Expired"

	ErrNoAccessRight  int    = 8150
	DescNoAccessRight string = "No Access Right"

	ErrLoginFailedUnKnown  int    = 8199
	DescLoginFailedUnKnown string = "Login Failed - UnKnown"

	//Bad Request
	ErrOthers  int    = 9999
	DescOthers string = "[9999]Unknown Error"

	//Request Error
	ErrInvalidFormat  int    = 9010
	DescInvalidFormat string = "Invalid Format"

	ErrInvalidParameter  int    = 9011
	DescInvalidParameter string = "Invalid Parameter"

	ErrInvalidStartEndDate  int    = 9012
	DescInvalidStartEndDate string = "Wrong Start or End Date"

	ErrInvalidDateFormat  int    = 9013
	DescInvalidDateFormat string = "Invalid Date Format, Please use YYYY-MM-DD Format"
	ErrInvalidJSONFormat  int    = 4

	ErrTextNotContainUniqueCode  int    = 2010
	DescTextNotContainUniqueCode string = "Text not contain UniqueCode"

	ErrSourceBlacklisted  int    = 2020
	DescSourceBlacklisted string = "Source MSISDN Blacklisted"

	ErrDestBlacklisted  int    = 2030
	DescDestBlacklisted string = "Destination MSISDN Blacklisted"

	ErrFormatNotValid  int    = 2040
	DescFormatNotValid string = "Message Format Not Valid"

	ErrUniqueCodeNotValid  int    = 2050
	DescUniqueCodeNotValid string = "UniqueCode Not Valid"

	ErrUniqueCodeUsed  int    = 2060
	DescUniqueCodeUsed string = "UniqueCode Used"

	ErrClaimCodeUsed  int    = 2061
	DescClaimCodeUsed string = "ClaimCode Used"

	ErrUnknownKeyword  int    = 2070
	DescUnknownKeyword string = "Unknown Keywords"

	ErrIgnoredMSISDN  int    = 2071
	DescIgnoredMSISDN string = "Ignored MSISDN"

	ErrDrawError  int    = 2080
	DescDrawError string = "Draw System Error"

	ErrPPOBFailed  int    = 2081
	DescPPOBFailed string = "PPOB Response Failed"

	ErrRequestThrottled  int    = 2082
	DescRequestThrottled string = "Request Throttled"

	ErrUniqueCodeSystemErrorValidate  int    = 2090
	DescUniqueCodeSystemErrorValidate string = "System Error When Validate Unique Code"

	ErrClaimCodeSystemErrorValidate  int    = 2091
	DescClaimCodeSystemErrorValidate string = "System Error When Validate Claim Code"

	ErrPrizeQtyLowerThanDistribute  int    = 2220
	DescPrizeQtyLowerThanDistribute string = "Prize qty lower than distribute"

	ErrUniqueCodeLocked  int    = 2221
	DescUniqueCodeLocked string = "Unique Code Locked"

	ErrClaimCodeLocked  int    = 2222
	DescClaimCodeLocked string = "Claim Code Locked"

	ErrWrongPram  int    = 2230
	DescWrongPram string = "Wrong Param"

	ErrEndCampaign  int    = 2910
	DescEndCampaign string = "End Campaign"
)
