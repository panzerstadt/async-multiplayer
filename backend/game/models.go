package game

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Email        string    `json:"email" gorm:"unique"`
	AuthProvider string    `json:"auth_provider"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

type Game struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	Name          string     `json:"name" gorm:"unique"`
	CreatorID     uuid.UUID  `json:"creator_id"`
	CurrentTurnID *uuid.UUID `json:"current_turn_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Players       []Player   `json:"players" gorm:"foreignKey:GameID"`
}

func (g *Game) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}

type Player struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	GameID    uuid.UUID `json:"game_id"`
	TurnOrder int       `json:"turn_order"`
}

func (p *Player) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

type Save struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	GameID     uuid.UUID `json:"game_id"`
	FilePath   string    `json:"file_path"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Save) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
