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

type Source interface {
	Add(migration migrations.Migration) error
}

type migrationOpts struct {
	skip   int
	source Source
}

// Option is a function that can be used to configure the Migration2 and Migration.
type Option func(opts *migrationOpts)

func defaultMigrationOpts() migrationOpts {
	return migrationOpts{
		skip:   1,
		source: migrations.DefaultSource,
	}
}

// WithSkip is an option to skip the number of callers from the stack trace. This is useful when you have a helper
// function that calls Migration or Migration2.
func WithSkip(skip int) Option {
	return func(opts *migrationOpts) {
		opts.skip = skip
	}
}

// WithSource is an option to set a custom source to auto-register the migration.
func WithSource(source Source) Option {
	return func(opts *migrationOpts) {
		opts.source = source
	}
}

// Migration2 is a helper function to create a new migration based on the filename of the caller.
// Eg: if you have a file called 1234567890_some_description.go, the migration ID will be 1234567890 and the description
// will be "some description".
//
// Migration2 can panic if the migration cannot be added to the source.
func Migration2(do, undo func(ctx context.Context) error, opts ...Option) migrations.Migration {
	o := defaultMigrationOpts()
	for _, opt := range opts {
		opt(&o)
	}
	_, file, _, ok := runtime.Caller(o.skip)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrInvalidFilename, path.Base(file)))
	}
	m := createMigration(file, do, undo)
	if o.source != nil {
		err := o.source.Add(m)
		if err != nil {
			panic(err)
		}
	}
	return m
}

// Migration is a helper function to create a new forward migration based on the filename of the caller. The
// difference between this and Migration2 is that this doesn't need the undo function.
//
// Migration can panic if the migration cannot be added to the source.
func Migration(do func(ctx context.Context) error, opts ...Option) migrations.Migration {
	o := defaultMigrationOpts()
	for _, opt := range opts {
		opt(&o)
	}
	_, file, _, ok := runtime.Caller(o.skip)
	if !ok {
		panic(fmt.Errorf("%w: %s", ErrInvalidFilename, path.Base(file)))
	}
	m := createMigration(file, do, nil)
	if o.source != nil {
		err := o.source.Add(m)
		if err != nil {
			panic(err)
		}
	}
	return m
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
