package response

type DebugEchoResponse struct {
	Message string `json:"message"`
}

type DebugTokenResponse struct {
	IDToken string `json:"id_token" example:"eyJhbGciOiJSUzI1NiIs..."`
	UID     string `json:"uid" example:"hTFudCor2wT3Bab570LvOm2X6z73"`
}
