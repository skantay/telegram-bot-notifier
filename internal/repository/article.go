package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/skantay/notifier/internal/model"
)

type ArticleRepository struct {
	db *sqlx.DB
}

func NewArticlerepository(db *sqlx.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (s *ArticleRepository) Store(ctx context.Context, article model.Article) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx,
		`INSERT INTO articles(source_id, title, link, summary, published_at)
		VALUES($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.PublishedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *ArticleRepository) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]model.Article, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle

	if err := conn.SelectContext(
		ctx,
		&articles,
		`SELECT * FROM articles
		WHER posted_at IS NULL AND published_at >= $1::timestamp
		ORDER BY published_at DESC
		LIMIT $2`,
		since.UTC().Format(time.RFC3339),
		limit,
	); err != nil {
		return nil, err
	}

	return lo.Map(articles, func(article dbArticle, _ int) model.Article {
		return model.Article{
			ID:          article.ID,
			SourceID:    article.SourceID,
			Title:       article.Title,
			Link:        article.Link,
			Summary:     article.Summary,
			PostedAt:    article.PostedAt.Time,
			PublishedAt: article.PublishedAt,
			CreatedAt:   article.CreatedAt,
		}
	}), nil
}

func (s *ArticleRepository) MarkPosted(ctx context.Context, id int64) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`UPDATE articles
		 SET posted_at = $1::timestamp
		 WHERE id = $2`,
		time.Now().UTC().Format(time.RFC3339),
		id,
	); err != nil {
		return err
	}

	return nil
}

type dbArticle struct {
	ID          int64        `db:"id"`
	SourceID    int64        `db:"source_id"`
	Title       string       `db:"title"`
	Link        string       `db:"link"`
	Summary     string       `db:"summary"`
	PublishedAt time.Time    `db:"published_at"`
	CreatedAt   time.Time    `db:"created_at"`
	PostedAt    sql.NullTime `db:"posted_at"`
}
