package data

import "ScoreTableApi/internal/validator"

type Dto interface {
	Convert(v *validator.Validator) (resource Resource, aux any)
	Merge(v *validator.Validator, r Resource) (aux any)
	validate(v *validator.Validator)
}

type Resource interface {
	ToDto() Dto
}

type AnonymousDto struct{}

func (dto AnonymousDto) ToResource(_ *validator.Validator) (Resource, any) {
	return nil, nil
}
