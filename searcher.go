package stuber

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

// ErrServiceNotFound is returned when the service is not found.
var ErrServiceNotFound = errors.New("service not found")

// ErrMethodNotFound is returned when the method is not found.
var ErrMethodNotFound = errors.New("method not found")

// ErrStubNotFound is returned when the stub is not found.
var ErrStubNotFound = errors.New("stub not found")

// searcher is a struct that manages the storage of search results.
//
// It contains a mutex for concurrent access, a map to store and retrieve
// used stubs by their UUID, and a pointer to the storage struct.
type searcher struct {
	mu       sync.RWMutex // mutex for concurrent access
	stubUsed map[uuid.UUID]struct{}
	// map to store and retrieve used stubs by their UUID

	storage *storage // pointer to the storage struct
}

// newSearcher creates a new instance of the searcher struct.
//
// It initializes the stubUsed map and the storage pointer.
//
// Returns a pointer to the newly created searcher struct.
func newSearcher() *searcher {
	return &searcher{
		storage:  newStorage(),
		stubUsed: make(map[uuid.UUID]struct{}),
	}
}

// Result represents the result of a search operation.
//
// It contains two fields: found and similar. Found represents the exact
// match found in the search, while similar represents the most similar match
// found.
type Result struct {
	found   *Stub // The exact match found in the search
	similar *Stub // The most similar match found
}

// Found returns the exact match found in the search.
//
// Returns a pointer to the Stub struct representing the found match.
func (r *Result) Found() *Stub {
	return r.found
}

// Similar returns the most similar match found in the search.
//
// Returns a pointer to the Stub struct representing the similar match.
func (r *Result) Similar() *Stub {
	return r.similar
}

// upsert inserts the given stub values into the searcher. If a stub value
// already exists with the same key, it is updated.
//
// The function returns a slice of UUIDs representing the keys of the
// inserted or updated values.
func (s *searcher) upsert(values ...*Stub) []uuid.UUID {
	return s.storage.upsert(s.castToValue(values)...)
}

// del deletes the stub values with the given UUIDs from the searcher.
//
// Returns the number of stub values that were successfully deleted.
func (s *searcher) del(ids ...uuid.UUID) int {
	return s.storage.del(ids...)
}

// findByID retrieves the stub value associated with the given ID from the
// searcher.
//
// Returns a pointer to the Stub struct associated with the given ID, or nil
// if not found.
func (s *searcher) findByID(id uuid.UUID) *Stub {
	if v, ok := s.storage.findByID(id).(*Stub); ok {
		return v
	}

	return nil
}

// findBy retrieves all Stub values that match the given service and method
// from the searcher.
//
// Parameters:
// - service: The service field used to search for Stub values.
// - method: The method field used to search for Stub values.
//
// Returns:
// - []*Stub: The Stub values that match the given service and method, or nil if not found.
// - error: An error if the search fails.
func (s *searcher) findBy(service, method string) ([]*Stub, error) {
	// Retrieve all Stub values that match the given service and method from the storage.
	all, err := s.storage.findAll(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Cast the values to Stub pointers and return.
	return s.castToStub(all), nil
}

// clear resets the searcher.
//
// It clears the stubUsed map and calls the storage clear method.
func (s *searcher) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the stubUsed map.
	s.stubUsed = make(map[uuid.UUID]struct{})

	// Clear the storage.
	s.storage.clear()
}

// all returns all Stub values stored in the searcher.
//
// Returns:
// - []*Stub: The Stub values stored in the searcher.
func (s *searcher) all() []*Stub {
	// Cast the values to Stub pointers and return.
	return s.castToStub(s.storage.values())
}

// used returns all Stub values that have been used by the searcher.
//
// Returns:
// - []*Stub: The Stub values that have been used by the searcher.
func (s *searcher) used() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Retrieve all Stub values with keys in the stubUsed map.
	return s.castToStub(s.storage.findByIDs(maps.Keys(s.stubUsed)...))
}

// unused returns all Stub values that have not been used by the searcher.
//
// Returns:
// - []*Stub: The Stub values that have not been used by the searcher.
func (s *searcher) unused() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize an empty slice to store the results.
	results := make([]*Stub, 0, len(s.all()))

	// Iterate over all Stub values.
	for _, stub := range s.all() {
		// Check if the stub has not been used.
		if _, ok := s.stubUsed[stub.ID]; !ok {
			// Add the stub to the results.
			results = append(results, stub)
		}
	}

	// Return the results.
	return results
}

// find retrieves the Stub value associated with the given Query from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) find(query Query) (*Result, error) {
	// Check if the Query has an ID field.
	if query.ID != nil {
		// Search for the Stub value with the given ID.
		return s.searchByID(query.Service, query.Method, query)
	}

	// Search for the Stub value with the given service and method.
	return s.search(query)
}

// searchByID retrieves the Stub value associated with the given ID from the searcher.
//
// Parameters:
// - service: The service field used to search for the Stub value.
// - method: The method field used to search for the Stub value.
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) searchByID(service, method string, query Query) (*Result, error) {
	// Check if the given service and method are valid.
	_, err := s.storage.posByN(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Search for the Stub value with the given ID.
	if found := s.findByID(*query.ID); found != nil {
		// Mark the Stub value as used.
		s.mark(query, *query.ID)

		// Return the found Stub value.
		return &Result{found: found}, nil
	}

	// Return an error if the Stub value is not found.
	return nil, ErrServiceNotFound
}

// search retrieves the Stub value associated with the given Query from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) search(query Query) (*Result, error) {
	// Find all Stub values with the given service and method.
	stubs, err := s.findBy(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Initialize variables to store the found and similar Stub values.
	var (
		found       *Stub
		foundRank   float64
		similar     *Stub
		similarRank float64
	)

	// Iterate over the found Stub values.
	for _, stub := range stubs {
		// Calculate the rank of the current Stub value.
		current := rankMatch(query, stub)

		// Update the similar Stub value if the current rank is higher.
		if current > similarRank {
			similar = stub
			similarRank = current
		}

		// Update the found Stub value if the current Stub value matches the query and has a higher rank.
		if match(query, stub) && current > foundRank {
			found = stub
			foundRank = current
		}
	}

	// If a found Stub value is found, mark it as used and return it.
	if found != nil {
		s.mark(query, found.ID)

		return &Result{found: found}, nil
	}

	// If no found Stub value is found, return the similar Stub value.
	if similar == nil {
		return nil, ErrStubNotFound
	}

	return &Result{found: nil, similar: similar}, nil
}

// mark marks the given Stub value as used in the searcher.
//
// If the query's RequestInternal flag is set, the mark is skipped.
//
// Parameters:
// - query: The query used to mark the Stub value.
// - id: The UUID of the Stub value to mark.
func (s *searcher) mark(query Query, id uuid.UUID) {
	// If the query's RequestInternal flag is set, skip the mark.
	if query.RequestInternal() {
		return
	}

	// Lock the mutex to ensure concurrent access.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark the Stub value as used by adding it to the stubUsed map.
	s.stubUsed[id] = struct{}{}
}

// castToValue converts a slice of *Stub values to a slice of Value interface{}.
//
// Parameters:
// - values: A slice of *Stub values to convert.
//
// Returns:
// - A slice of Value interface{} containing the converted values.
func (s *searcher) castToValue(values []*Stub) []Value {
	result := make([]Value, len(values))
	for i, v := range values {
		result[i] = v
	}

	return result
}

// castToStub converts a slice of Value interface{} to a slice of *Stub.
//
// Parameters:
// - values: A slice of Value interface{} to convert.
//
// Returns:
// - A slice of *Stub containing the converted values.
func (s *searcher) castToStub(values []Value) []*Stub {
	ret := make([]*Stub, 0, len(values))

	for _, v := range values {
		if s, ok := v.(*Stub); ok {
			ret = append(ret, s)
		}
	}

	return ret
}

// wrap wraps an error with specific error types.
//
// Parameters:
// - err: The error to wrap.
//
// Returns:
// - The wrapped error.
func (s *searcher) wrap(err error) error {
	if errors.Is(err, ErrLeftNotFound) {
		return ErrServiceNotFound
	}

	if errors.Is(err, ErrRightNotFound) {
		return ErrMethodNotFound
	}

	return err
}
