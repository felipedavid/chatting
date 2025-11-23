package main

import (
	"context"

	"log/slog"

	"github.com/felipedavid/chatting/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func run() error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@127.0.0.1:5432/chatting")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	queries := storage.New(conn)

	user, err := queries.CreateUser(ctx, storage.CreateUserParams{
		PhoneNumber: "+1234567890",
		DisplayName: pgtype.Text{String: "Felipe David", Valid: true},
		About:       pgtype.Text{String: "Hello there!!!", Valid: true},
	})
	if err != nil {
		return err
	}

	slog.Info("New user", "data", user)

	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
