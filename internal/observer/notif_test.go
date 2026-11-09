package observer

import (
	"database/sql"
	"final/internal/structs"
	"testing"

	_ "modernc.org/sqlite"
)

func testPostReactionTables(t *testing.T) *Repo {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		db.Close()
	})

	schema := `CREATE TABLE IF NOT EXISTS post_reactions (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    value INTEGER NOT NULL CHECK (value in (-1, 1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id)    
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}

	return NewRepo(db)
}

func TestSavingIntoPostReactions(t *testing.T) {
	repo := testPostReactionTables(t)

	reactionPost := &structs.PostReaction{
		UserID: 1,
		PostID: 1,
	}

	err := repo.SavePostReactions(reactionPost.UserID, reactionPost.PostID, structs.ReactionLike)
	if err != nil {
		t.Fatal(err)
	}

	var rows int
	var value structs.Reaction
	err = repo.db.QueryRow(
		`SELECT COUNT(*), COALESCE(MAX(value), 0) FROM post_reactions WHERE user_id = ? AND post_id = ?`,
		reactionPost.UserID, reactionPost.PostID).Scan(&rows, &value)
	if err != nil {
		t.Fatal(err)
	}
	if rows != 1 || value != structs.ReactionLike {
		t.Fatalf("rows=%d value=%d, want rows=1 value=%d", rows, value, structs.ReactionLike)
	}
}

func TestGetPostReaction(t *testing.T) {
	repo := testPostReactionTables(t)

	for i := 1; i < 10; i++ {
		err := repo.SavePostReactions(i, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
	}
	err := repo.SavePostReactions(10, 1, -1)
	if err != nil {
		t.Fatal(err)
	}
	err = repo.SavePostReactions(11, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	summary, err := repo.GetPostReactionSummary(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if summary.LikeCount != 9 {
		t.Fatalf("got %d, want 9", summary.LikeCount)
	}
	if summary.DislikeCount != 1 {
		t.Fatalf("got %d, want 1", summary.DislikeCount)
	}
	if summary.UserReaction != structs.ReactionLike {
		t.Fatalf("got %d, want %d", summary.UserReaction, structs.ReactionLike)
	}
}
