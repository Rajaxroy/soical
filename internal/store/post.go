package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comment   []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int `json:"comments_count"`
}
type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) GetUserFeed(ctx context.Context, userID int64) ([]PostWithMetaData, error) {
	query := `
		select 
			p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, u.username, COUNT (c.id) as comment_count
		from
			posts p
		left join comments c on
			c.post_id = p.id
		left join users u on
			p.user_id = u.id
		join followers f on
			f.follower_id = p.user_id or p.user_id = $1
		where
			f.user_id = $1 or p.user_id = $1
		group by p.id, u.username
		order by created_at desc;
	`
	row, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer row.Close()

	var feed []PostWithMetaData
	for row.Next() {
		var p PostWithMetaData
		err := row.Scan(
			&p.ID,
			&p.UserId,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentCount,
		)

		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	return err
}

func (s *PostsStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at, tags, version
		FROM posts
		where id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}
	return &post, nil

}

func (s *PostsStore) Delete(ctx context.Context, postID int64) error {
	query := `DELETE FROM posts WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrorNotFound
	}
	return nil
}

func (s *PostsStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx, query, post.Title, post.Content, post.ID, post.Version,
	).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrorNotFound
		default:
			return err
		}
	}

	return nil
}
