package repository

import (
	"database/sql"
	"microblogging/model"
)

type PostRepository interface {
	Save(post *model.Post) error
	GetTimeline(userID string) ([]model.Post, error)
	FollowUser(followerID, followeeID string) error
	GetFollowees(userID string) ([]string, error)
}

type postRepo struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepo{db: db}
}

// Save persists the given post to the underlying database.
func (r *postRepo) Save(post *model.Post) error {
	_, err := r.db.Exec(`INSERT INTO posts (id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4)`,
		post.ID, post.UserID, post.Content, post.CreatedAt)
	return err
}

// GetTimeline returns a slice of posts that the user with the given ID has seen
// in their timeline, ordered in descending order of creation time.
func (r *postRepo) GetTimeline(userID string) ([]model.Post, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.user_id, p.content, p.created_at
		FROM posts p
		JOIN follows f ON f.following_id = p.user_id
		WHERE f.follower_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Content, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

// FollowUser adds a follow relationship between the given follower and followee.
// If such a relationship already exists, the call does nothing.
func (r *postRepo) FollowUser(followerID, followeeID string) error {
	query := `
		INSERT INTO follows (follower_id, followee_id) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(query, followerID, followeeID)
	return err
}

// GetFollowees returns a slice of followee IDs that the user with the given ID
// follows.
func (r *postRepo) GetFollowees(userID string) ([]string, error) {
	query := `SELECT followee_id FROM follows WHERE follower_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followees []string
	for rows.Next() {
		var followeeID string
		if err := rows.Scan(&followeeID); err != nil {
			return nil, err
		}
		followees = append(followees, followeeID)
	}
	return followees, nil
}
