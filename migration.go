package migrationsfnc

import (
	"context"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/jamillosantos/migrations"
)

var (
	ErrInvalidFilename              = errors.New("invalid filename")
	ErrMigrationDescriptionRequired = errors.New("migration description is required")
)

// getMigrationID will return the migration ID based on the filename from the function it was called from removing the
// file extension.
// Example:
// - given 4829481231293_some_description.go returns 4829481231293 and "some description"
func getMigrationInfo(file string) (string, string, error) {
	s := strings.Split(path.Base(strings.TrimSuffix(file, path.Ext(file))), "_")
	var description string
	if len(s) < 2 {
		return s[0], "", fmt.Errorf("%w: %s", ErrMigrationDescriptionRequired, file)
	}
	description = strings.Join(s[1:], " ")
	return s[0], description, nil
}

func NewMigration(description string, do, undo func(ctx context.Context) error) (migrations.Migration, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFilename, path.Base(file))
	}
	id, description, err := getMigrationInfo(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration ID: %w", err)
	}
	return migrations.NewMigration(id, description, do, undo), nil
}

type migrationOpts struct {
	skip int
}

// Option is a function that can be used to configure the Migration2 and Migration.
type Option func(opts *migrationOpts)

func defaultMigrationOpts() migrationOpts {
	return migrationOpts{
		skip: 1,
	}
}

// Migration2 is a helper function to create a new migration based on the filename of the caller.
// Eg: if you have a file called 1234567890_some_description.go, the migration ID will be 1234567890 and the description
// will be "some description".
func Migration2(description string, do, undo func(ctx context.Context) error, opts ...Option) migrations.Migration {
	o := defaultMigrationOpts()
	for _, opt := range opts {
		opt(&o)
	}
	_, file, _, ok := runtime.Caller(o.skip)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrInvalidFilename, path.Base(file)))
	}
	return createMigration(file, do, undo)
}

// Migration is a helper function to create a new forward migration based on the filename of the caller. The
// difference between this and Migration2 is that this doesn't need the undo function.
func Migration(description string, do func(ctx context.Context) error, opts ...Option) migrations.Migration {
	o := defaultMigrationOpts()
	for _, opt := range opts {
		opt(&o)
	}
	_, file, _, ok := runtime.Caller(o.skip)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrInvalidFilename, path.Base(file)))
	}
	return createMigration(file, do, nil)
}

func createMigration(file string, do, undo func(ctx context.Context) error) migrations.Migration {
	id, description, err := getMigrationInfo(file)
	if err != nil {
		panic(fmt.Errorf("failed to get migration ID: %w", err))
	}
	m, err := migrations.NewMigration(id, description, do, undo), nil
	if err != nil {
		panic(err)
	}
	return m
}
