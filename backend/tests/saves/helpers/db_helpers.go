package helpers

import (
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
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

// Game represents a game session
type Game struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
	Name          string     `json:"name" gorm:"unique"`
	CreatorID     uuid.UUID  `json:"creator_id"`
	CurrentTurnID *uuid.UUID `json:"current_turn_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (g *Game) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}

// Player represents a player in a game
type Player struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `json:"user_id"`
	GameID    uuid.UUID `json:"game_id"`
	TurnOrder int       `json:"turn_order"`
}

func (p *Player) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// Save represents a save file
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

// CreateFakeUser creates a fake user in the test database
func CreateFakeUser(db *gorm.DB) *User {
	user := &User{
		Email:        gofakeit.Email(),
		AuthProvider: gofakeit.RandomString([]string{"google", "github", "discord"}),
		CreatedAt:    gofakeit.DateRange(time.Now().AddDate(0, -6, 0), time.Now()),
	}
	
	db.Create(user)
	return user
}

// CreateFakeGame creates a fake game in the test database
func CreateFakeGame(db *gorm.DB, creatorID uuid.UUID) *Game {
	game := &Game{
		Name:      gofakeit.Gamertag() + " " + gofakeit.RandomString([]string{"Adventure", "Quest", "Campaign", "Mission"}),
		CreatorID: creatorID,
		CreatedAt: gofakeit.DateRange(time.Now().AddDate(0, -3, 0), time.Now()),
		UpdatedAt: time.Now(),
	}
	
	db.Create(game)
	return game
}

// CreateFakePlayer creates a fake player in the test database
func CreateFakePlayer(db *gorm.DB, userID, gameID uuid.UUID, turnOrder int) *Player {
	player := &Player{
		UserID:    userID,
		GameID:    gameID,
		TurnOrder: turnOrder,
	}
	
	db.Create(player)
	return player
}

// CreateFakeSave creates a fake save file record in the test database
func CreateFakeSave(db *gorm.DB, gameID, uploadedBy uuid.UUID, filePath string) *Save {
	save := &Save{
		GameID:     gameID,
		FilePath:   filePath,
		UploadedBy: uploadedBy,
		CreatedAt:  gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now()),
	}
	
	db.Create(save)
	return save
}

// CreateFakeGameWithPlayers creates a complete fake game scenario with multiple players
func CreateFakeGameWithPlayers(db *gorm.DB, playerCount int) (*Game, []*User, []*Player) {
	// Create creator user
	creator := CreateFakeUser(db)
	
	// Create game
	game := CreateFakeGame(db, creator.ID)
	
	// Create additional users and players
	users := []*User{creator}
	players := []*Player{}
	
	// Create creator as first player
	creatorPlayer := CreateFakePlayer(db, creator.ID, game.ID, 0)
	players = append(players, creatorPlayer)
	
	// Create additional players
	for i := 1; i < playerCount; i++ {
		user := CreateFakeUser(db)
		player := CreateFakePlayer(db, user.ID, game.ID, i)
		users = append(users, user)
		players = append(players, player)
	}
	
	return game, users, players
}

// CleanupTestData removes all test data from the database
func CleanupTestData(db *gorm.DB) {
	db.Exec("DELETE FROM saves")
	db.Exec("DELETE FROM players")
	db.Exec("DELETE FROM games")
	db.Exec("DELETE FROM users")
}

// SetupTestDB sets up a clean test database with all tables
func SetupTestDB(db *gorm.DB) error {
	// Drop and recreate tables to ensure clean state
	db.Migrator().DropTable(&Save{}, &Player{}, &Game{}, &User{})
	return db.AutoMigrate(&User{}, &Game{}, &Player{}, &Save{})
}
