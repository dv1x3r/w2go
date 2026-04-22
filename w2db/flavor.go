package w2db

import "github.com/huandu/go-sqlbuilder"

var defaultFlavor = sqlbuilder.DefaultFlavor

func SetFlavor(flavor sqlbuilder.Flavor) {
	defaultFlavor = flavor
}
