package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/kseilons/messenger-backend/internal/models"
)

// UserRepository interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error)
	GetOnlineUsers(ctx context.Context) ([]*models.User, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger *slog.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, display_name, avatar_url, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.DisplayName, user.AvatarURL, user.Status)

	if err != nil {
		r.logger.Error("Failed to create user", "error", err, "user_id", user.ID)
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created", "user_id", user.ID, "username", user.Username)
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.AvatarURL, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get user by ID", "error", err, "user_id", id)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.AvatarURL, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get user by username", "error", err, "username", username)
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.AvatarURL, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get user by email", "error", err, "email", email)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, display_name = $4, avatar_url = $5, status = $6, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.DisplayName, user.AvatarURL, user.Status)

	if err != nil {
		r.logger.Error("Failed to update user", "error", err, "user_id", user.ID)
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("User updated", "user_id", user.ID)
	return nil
}

// UpdateStatus updates user status
func (r *userRepository) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) error {
	query := `
		UPDATE users
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, status)
	if err != nil {
		r.logger.Error("Failed to update user status", "error", err, "user_id", userID, "status", status)
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("User status updated", "user_id", userID, "status", status)
	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", "error", err, "user_id", id)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("User deleted", "user_id", id)
	return nil
}

// Search searches for users by query
func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.User, error) {
	sqlQuery := `
		SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE username ILIKE $1 OR display_name ILIKE $1 OR email ILIKE $1
		ORDER BY username
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, limit, offset)
	if err != nil {
		r.logger.Error("Failed to search users", "error", err, "query", query)
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.DisplayName,
			&user.AvatarURL, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan user", "error", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// GetOnlineUsers retrieves all online users
func (r *userRepository) GetOnlineUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, created_at, updated_at
		FROM users
		WHERE status = 'online'
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get online users", "error", err)
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.DisplayName,
			&user.AvatarURL, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan online user", "error", err)
			return nil, fmt.Errorf("failed to scan online user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate online users: %w", err)
	}

	return users, nil
}
