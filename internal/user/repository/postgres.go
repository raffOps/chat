package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/raffops/chat/internal/errs"
	model "github.com/raffops/chat/internal/models"
	"log"
	"net/http"
)

type PostgresRepository struct {
	db *sql.DB
}

// fetchUser is a helper method to create a 'model.User' object
func (p PostgresRepository) fetchUser(id, name, password, role sql.NullString,
	createdAt, updatedAt, deleteAt sql.NullTime) model.User {
	return model.User{
		Id:        id.String,
		Name:      name.String,
		Password:  password.String,
		Role:      role.String,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
		DeletedAt: deleteAt.Time,
	}
}

// Get retrieves a user from the Postgres database by their name or id.
// It returns a model.User object and an error.
// If the user is not found, it returns a 404 error.
// If there is an internal server error, it returns a 500 error.
// Otherwise, it returns the user and a nil error.
func (p PostgresRepository) GetUser(ctx context.Context, key, value string) (model.User, *errs.Err) {
	var id, username, password, role sql.NullString
	var createdAt, updatedAt, deleteAt sql.NullTime
	queryString := `
			SELECT 
			    id,
				name, 
				password,
				role,
			   	created_at,
			   	updated_at,
			   	deleted_at
			FROM public.user WHERE %s = $1`

	queryString = fmt.Sprintf(queryString, key)
	err := p.db.QueryRowContext(ctx, queryString, value).
		Scan(&id, &username, &password, &role, &createdAt, &updatedAt, &deleteAt)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return model.User{}, &errs.Err{Message: "no user found", Code: http.StatusNotFound}
	case err != nil:
		log.Printf("query error: %v\n", err)
		return model.User{}, &errs.Err{Message: "internal server error", Code: http.StatusInternalServerError}
	default:
		log.Printf("username is %s, account created on %s\n", username.String, createdAt.Time)
		return p.fetchUser(id, username, password, role, createdAt, updatedAt, deleteAt), nil
	}
}

func (p PostgresRepository) ListUser(ctx context.Context) ([]model.User, *errs.Err) {
	var users []model.User
	queryString := `SELECT id, name, password, role, created_at, updated_at, deleted_at FROM public.user`
	rows, err := p.db.QueryContext(ctx, queryString)
	if err != nil {
		log.Printf("query error: %v\n", err)
		return nil, &errs.Err{Message: "internal server error", Code: http.StatusInternalServerError}
	}
	defer rows.Close()
	for rows.Next() {
		var id, username, password, role sql.NullString
		var createdAt, updatedAt, deleteAt sql.NullTime
		err := rows.Scan(&id, &username, &password, &role, &createdAt, &updatedAt, &deleteAt)
		if err != nil {
			log.Printf("scan error: %v\n", err)
			return nil, &errs.Err{Message: "internal server error", Code: http.StatusInternalServerError}
		}
		users = append(users, p.fetchUser(id, username, password, role, createdAt, updatedAt, deleteAt))
	}
	return users, nil
}

func (p PostgresRepository) CreateUser(ctx context.Context, user model.User) (model.User, *errs.Err) {
	var id, name, password, role sql.NullString
	var createdAt sql.NullTime
	createUserQuery := `INSERT INTO public."user" (name, password, role) VALUES ($1, $2, $3) 
                                           RETURNING id, name, password, role, created_at`

	err := p.db.QueryRowContext(ctx, createUserQuery, user.Name, user.Password, user.Role).
		Scan(&id, &name, &password, &role, &createdAt)

	if err != nil {
		log.Println("error creating user: ", err)
		return model.User{}, &errs.Err{Message: "internal server error", Code: 500}
	}
	return p.fetchUser(id, name, password, role, createdAt, sql.NullTime{}, sql.NullTime{}), nil
}

func NewPostgresRepo(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

//func main() {
//	postgresConn := database.GetPostgresConn()
//	defer postgresConn.Close()
//	repo := NewPostgresRepo(postgresConn)
//	user, err := repo.CreateUser(context.Background(), model.User{Name: "test", Password: "password"})
//	if err != nil {
//		println(err)
//	}
//	fmt.Print(user)
//	get, err := repo.GetUser(context.Background(), "test")
//	if err != nil {
//		return
//	}
//	fmt.Print(get)
//}
