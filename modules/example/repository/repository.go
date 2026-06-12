// Package repository is the data-access layer for the Example module.
// All DB queries live here. Services must never call the DB directly.
package repository

// Repositories bundles all Example repositories.
// Pass the whole struct to service.NewServices so adding a repo never changes signatures.
type Repositories struct {
	Example *ExampleRepository
}

// NewRepositories constructs all repositories.
// Replace the comment with *ent.Client once your schema is generated:
//
//	func NewRepositories(db *ent.Client) *Repositories {
//	    return &Repositories{ Example: NewExampleRepository(db) }
//	}
func NewRepositories( /* db *ent.Client */ ) *Repositories {
	return &Repositories{
		Example: NewExampleRepository(),
	}
}

// Example is the read-model returned by repository queries.
// Replace with the generated ent.Example type once Ent is wired.
type Example struct {
	ID   int
	Name string
}

// ExampleRepository handles all Example persistence operations.
type ExampleRepository struct {
	// db *ent.Client
}

// NewExampleRepository constructs a ExampleRepository.
func NewExampleRepository( /* db *ent.Client */ ) *ExampleRepository {
	return &ExampleRepository{}
}

// Create persists a new record and returns its generated ID.
func (r *ExampleRepository) Create(name string) (int, error) {
	// TODO: r.db.Example.Create().SetName(name).Save(ctx)
	return 0, nil
}

// List returns all records.
func (r *ExampleRepository) List() ([]Example, error) {
	// TODO: r.db.Example.Query().All(ctx)
	return []Example{}, nil
}

// FindByID returns one record or nil if not found.
func (r *ExampleRepository) FindByID(id int) (*Example, error) {
	// TODO: r.db.Example.Get(ctx, id)
	return nil, nil
}

// Update persists changes to an existing record.
func (r *ExampleRepository) Update(id int, name string) error {
	// TODO: r.db.Example.UpdateOneID(id).SetName(name).Exec(ctx)
	return nil
}

// Delete hard-deletes a record.
func (r *ExampleRepository) Delete(id int) error {
	// TODO: r.db.Example.DeleteOneID(id).Exec(ctx)
	return nil
}
