package models

func All() []any {
	return []any{
		&Account{},
		&Profile{},
		// ... TODO: implement other models
	}
}
