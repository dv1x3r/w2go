package w2db

import "github.com/huandu/go-sqlbuilder"

var defaultFlavor = sqlbuilder.DefaultFlavor

// SetFlavor sets the package default SQL dialect for helpers that do not receive an explicit Flavor option.
func SetFlavor(flavor sqlbuilder.Flavor) {
	defaultFlavor = flavor
}
