package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound        = errors.New("record not found.")
	QueryTimeoutDuration = time.Second * 5
	ErrConflict          = errors.New("record already exists.")
)

type Storage struct {
	Posts interface {
		GetByID(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64) ([]PostWithMetaData, error)
	}

	Users interface {
		GetByID(context.Context, int64) (*User, error)
		Create(context.Context, *User) error
	}

	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostID(context.Context, int64) ([]Comment, error)
	}

	Followers interface {
		Follow(ctx context.Context, followerId int64, userId int64) error
		Unfollow(ctx context.Context, followerId int64, userId int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostsStore{db},
		Users:     &UserStore{db},
		Comments:  &CommentStore{db},
		Followers: &FollowerStore{db},
	}
}
