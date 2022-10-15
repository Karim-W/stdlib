package stdlib


type AuthProvider interface{
	GetAuthHeader() string
}