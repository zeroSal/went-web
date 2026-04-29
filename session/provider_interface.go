package session

import "github.com/zeroSal/went-web/user"

type ProviderInterface interface {
	Load(credential any) (user.Interface, error)
}
