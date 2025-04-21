package repository

import (
	"fmt"
	"microblogging/model"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type DBConnector struct {
	DB     *sqlx.DB
	Logger *zap.Logger
}

func (r *DBConnector) NewPostRepository() PostRepository {
	return &postRepo{db: r.DB, logger: r.Logger}
}

func (r *DBConnector) Save(post *model.Post) (uuid.UUID, error) {
	var postID uuid.UUID
	now := time.Now().UTC()

	const insertQuery = `
		INSERT INTO posts (user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.DB.QueryRow(insertQuery, post.UserID, post.Content, now, now).Scan(&postID)
	if err != nil {
		r.Logger.Error("Error inserting post", zap.Error(err))
		return uuid.Nil, err
	}
	// Update the user's last post asynchronously
	r.updateUserLastPostAsync(postID, post.UserID, now)

	r.Logger.Sugar().Infow("Post saved", "post_id", postID.String())
	return postID, nil
}

func (r *DBConnector) UpdatePostPut(post model.CreatePostRequest) error {
	now := time.Now().UTC()
	postUUID, err := uuid.Parse(post.PostID)
	if err != nil {
		r.Logger.Error("Invalid post_id UUID", zap.Error(err))
		return model.ErrInvalidUUID
	}

	if err := r.existPost(postUUID, post.UserID); err != nil {
		return model.ErrPostNotFound
	}

	const updateQuery = `
		UPDATE posts
		SET content = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4;
	`
	_, err = r.DB.Exec(updateQuery, post.Content, now, post.PostID, post.UserID)
	if err != nil {
		r.Logger.Error("Error updating post", zap.Error(err))
		return err
	}

	// Update the user's last post asynchronously
	r.updateUserLastPostAsync(postUUID, post.UserID, now)
	r.Logger.Sugar().Infow("Post updated", "post_id", post.PostID)
	return err
}

func (r *DBConnector) GetTimeline(info model.TimelineRequest) (model.TimelineResponse, error) {
	var posts model.TimelineResponse
	// if info.Before is not set, it previously used a default value of 3 days from now
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at
		FROM posts p
		JOIN follows f ON f.followee_id = p.user_id
		WHERE f.follower_id = $1
	    AND f.is_active = TRUE
		AND p.created_at < $2
		ORDER BY p.created_at DESC
		LIMIT $3
	`
	err := r.DB.Select(&posts.Posts, query, info.UserID, info.Before, info.Limit)

	if err != nil {
		r.Logger.Sugar().Errorw("Error getting timeline", "error", err, "user_id", info.UserID, "before", info.Before, "limit", info.Limit)
		return model.TimelineResponse{}, err
	}
	return posts, nil
}

func (r *DBConnector) FollowUser(followerID, followeeID string) error {
	exists, err := r.checkUsers(followerID, followeeID)
	if !exists {
		return err
	}

	query := `
		INSERT INTO follows (follower_id, followee_id) 
		VALUES ($1, $2) 
		ON CONFLICT (follower_id, followee_id)
		DO UPDATE SET is_active = TRUE;
	`
	_, err = r.DB.Exec(query, followerID, followeeID)
	if err != nil {
		r.Logger.Sugar().Errorw("Error following user", "error", err, "user_id", followerID, "followee_id", followeeID)
	}
	return err
}

func (r *DBConnector) checkUsers(followerID string, followeeID string) (bool, error) {
	var (
		g  errgroup.Group
		wg sync.WaitGroup

		followerErr, followeeErr       error
		existsFollower, existsFollowee bool
	)

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()
		existsFollower, followerErr = r.existUser(followerID)
		if followerErr != nil || !existsFollower {
			return fmt.Errorf("%s: %s", model.ErrUserNotFound.Error(), followerID)
		}
		return nil
	})

	wg.Add(1)
	g.Go(func() error {
		defer wg.Done()
		existsFollowee, followeeErr = r.existUser(followeeID)
		if followeeErr != nil || !existsFollowee {
			return fmt.Errorf("%s: %s", model.ErrUserNotFound.Error(), followeeID)
		}
		return nil
	})

	wg.Wait()

	if followerErr != nil {
		return false, followerErr
	}
	if followeeErr != nil {
		return false, followeeErr
	}

	return true, nil
}

func (r *DBConnector) UnfollowUser(followerID, followeeID string) error {
	exists, err := r.checkUsers(followerID, followeeID)
	if !exists {
		return err
	}

	query := `
		UPDATE follows
		SET is_active = FALSE
		WHERE follower_id = $1 AND followee_id = $2;
	`
	_, err = r.DB.Exec(query, followerID, followeeID)
	if err != nil {
		r.Logger.Sugar().Errorw("Error unfollowing user", "error", err, "user_id", followerID, "followee_id", followeeID)
	}
	return err
}

func (r *DBConnector) GetFollowees(userID string, limit int) ([]string, error) {
	var followees []string
	query := `SELECT followee_id
			  FROM follows 
			  WHERE follower_id = $1
			  LIMIT $2`
	err := r.DB.Select(&followees, query, userID, limit)
	if err != nil {
		r.Logger.Error("Error getting followees", zap.Error(err))
		return nil, err
	}
	r.Logger.Sugar().Infow("Got followees info", "user_id", userID)
	return followees, nil
}

func (r *DBConnector) CreateUser(userData model.CreateUserRequest) (uuid.UUID, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var userID uuid.UUID
	query := `
		INSERT INTO users (user_name, password, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	err := r.DB.QueryRow(query, userData.Name, userData.Password, userData.Email, now, now).Scan(&userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert user: %w", err)
	}
	r.Logger.Sugar().Infow("User created successfully", "user_id", userID.String())
	return userID, nil
}

func (r *DBConnector) GetUser(userID string) (model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = $1`
	if err := r.DB.Get(&user, query, userID); err != nil {
		r.Logger.Sugar().Errorw("Error getting user", "error", err, "user_id", userID)
		return model.User{}, err
	}
	r.Logger.Sugar().Infow("Got user info", "user_id", userID)
	return user, nil
}

func (r *DBConnector) DeleteUser(userID string) error {
	exists, err := r.existUser(userID)
	if err != nil || !exists {
		r.Logger.Error("User not found", zap.Error(err))
		return err
	}
	query := `DELETE FROM users WHERE id = $1`
	_, err = r.DB.Exec(query, userID)
	if err != nil {
		r.Logger.Error("Error deleting user", zap.Error(err))
		return err
	}
	r.Logger.Sugar().Info("User is deleted", "user_id", userID)
	return err
}
func (r *DBConnector) existUser(userID string) (bool, error) {
	var exists bool
	checkPostQuery := `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1);`
	err := r.DB.QueryRow(checkPostQuery, userID).Scan(&exists)
	if err != nil {
		r.Logger.Sugar().Errorw("Error checking if user exists", "error", err, "user_id", userID)
		return false, err
	}

	if !exists {
		r.Logger.Sugar().Errorw("user_id does not exist", "user_id", userID)
		return exists, model.ErrUserNotFound
	}
	r.Logger.Sugar().Infow("Existing user", "user_id", userID)
	return exists, nil
}
func (r *DBConnector) updateUserLastPostAsync(postID uuid.UUID, userID string, updatedAt time.Time) {
	go func() {
		const updateUserQuery = `
			UPDATE users
			SET last_post_id = $1, updated_at = $2
			WHERE id = $3;
		`
		if _, err := r.DB.Exec(updateUserQuery, postID, updatedAt, userID); err != nil {
			r.Logger.Error("Error updating user's last_post_id", zap.Error(err))
		}
		r.Logger.Sugar().Infow("User's last_post_id updated", "user_id", userID, "post_id", postID.String())
	}()
}

func (r *DBConnector) existPost(postID uuid.UUID, userID string) error {
	var exists bool
	checkPostQuery := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND user_id = $2);`
	err := r.DB.QueryRow(checkPostQuery, postID, userID).Scan(&exists)
	if err != nil {
		r.Logger.Error("Error checking if post exists", zap.Error(err))
		return err
	}

	if !exists {
		r.Logger.Sugar().Errorw("post_id does not exist for user_id", "post_id", postID.String(), "user_id", userID)
		return model.ErrPostNotFound
	}
	r.Logger.Sugar().Infow("Existing post", "post_id", postID.String(), "user_id", userID)
	return nil
}
