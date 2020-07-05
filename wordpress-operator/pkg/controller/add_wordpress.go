package controller

import (
	"github.com/renan-campos/wordpress-operator/pkg/controller/wordpress"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, wordpress.Add)
}
