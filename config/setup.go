package config

import (
	"context"
	"fmt"
	t "microblogging/model"
	d "microblogging/repository"
	srv "microblogging/server"
	"microblogging/service"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type flags struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	DBName   string `validate:"required"`
	SSLMode  string
}

// Setup
func Setup(ctx context.Context) (d.PostRepository, error) {
	// Get DB parameters from flags
	dbConfig, err := setupFlags()
	if err != nil {
		return nil, fmt.Errorf("could not get DB params: %w", err)
	}

	// Setup DB connection
	db, err := SetupDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("could not configure DB: %w", err)
	}

	// Setup logger
	logger, err := SetupLogger()
	if err != nil {
		return nil, fmt.Errorf("could not configure logger: %w", err)
	}

	// Return a PostRepository (DBConnector now implements PostRepository)
	return SetupRepository(db, logger), nil
}

func setupFlags() (t.DatabaseConfig, error) {
	args := flags{
		Host: os.Getenv("POSTGRES_HOST"),
		Port: func() int {
			port, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
			if err != nil {
				return 0
			}
			return port
		}(),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}

	v := validator.New()
	if err := v.Struct(args); err != nil {
		return t.DatabaseConfig{}, err
	}

	return t.DatabaseConfig{
		Host:     args.Host,
		Port:     args.Port,
		User:     args.User,
		Password: args.Password,
		DBName:   args.DBName,
		SSLMode:  args.SSLMode,
	}, nil
}

func SetupDB(config t.DatabaseConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetupLogger all necessary stuff to configure logger
func SetupLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

func SetupRepository(db *sqlx.DB, logger *zap.Logger) d.PostRepository {
	return &d.DBConnector{
		DB:     db,
		Logger: logger,
	}
}

func ServerSetup(svc service.BlogService) {
	s := srv.NewServer(context.Background(), svc)

	router := mux.NewRouter()
	api := router.PathPrefix("/V1").Subrouter()
	api.HandleFunc("/post", s.CreatePostHandler).Methods("POST")
	api.HandleFunc("/timeline", s.GetTimelineHandler).Methods("GET")
	api.HandleFunc("/follow", s.FollowUserHandler).Methods("POST")
	api.HandleFunc("/followees/{id}", s.GetFolloweesHandler).Methods("GET")

	port := ":8080"
	http.ListenAndServe(port, router)
}
