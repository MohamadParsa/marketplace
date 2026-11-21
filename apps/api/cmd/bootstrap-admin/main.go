package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"time"
)

func main() {
	phone := flag.String("phone", "", "first admin mobile number")
	flag.Parse()
	pass := os.Getenv("CHERRIVA_BOOTSTRAP_PASSWORD")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || len(pass) < 8 || len(pass) > 72 || len(*phone) != 11 || !strings.HasPrefix(*phone, "09") {
		fmt.Fprintln(os.Stderr, "valid live database, phone and bootstrap password environment variable required")
		os.Exit(1)
	}
	for _, r := range *phone {
		if r < '0' || r > '9' {
			fmt.Fprintln(os.Stderr, "invalid phone")
			os.Exit(1)
		}
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if e != nil {
		fmt.Fprintln(os.Stderr, "password hashing failed")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, e := pgxpool.New(ctx, dsn)
	if e != nil {
		fmt.Fprintln(os.Stderr, "database connection failed")
		os.Exit(1)
	}
	defer db.Close()
	repo, e := postgresrepo.NewLive(ctx, db)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	e = repo.Update(ctx, func(w *backoffice.Workspace) error {
		for _, u := range w.Users {
			if u.Role == identity.RoleSuperAdmin {
				return fmt.Errorf("an administrator already exists; use the authorized panel")
			}
		}
		buf := make([]byte, 24)
		if _, e := rand.Read(buf); e != nil {
			return e
		}
		w.Users = append(w.Users, backoffice.User{ID: hex.EncodeToString(buf), Phone: *phone, Name: "مدیر", Role: identity.RoleSuperAdmin, Active: true, AccessVersion: 1, PasswordHash: string(hash)})
		return nil
	})
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Println("First administrator created; no credentials printed.")
}
