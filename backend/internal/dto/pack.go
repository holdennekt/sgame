package dto

import "github.com/holdennekt/sgame/internal/domain"

type CreatePackDTO struct {
	UserId  string
	PackDTO *domain.PackDTO
}

type GetPackByIdDTO struct {
	UserId string
	Id     string
}

type GetPackByRoundsChecksumDTO struct {
	RoundsChecksum []byte
	IgnoreId       string
}

type GetPacksDTO struct {
	UserId string
	SearchDTO
}

type UpdatePackDTO struct {
	UserId  string
	Id      string
	PackDTO *domain.PackDTO
}

type DeletePackDTO struct {
	UserId string
	Id     string
}
