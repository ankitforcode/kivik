package driver

import (
	"context"
	"encoding/json"
	"io"
)

// Driver is the interface that must be implemented by a database driver.
type Driver interface {
	// NewClient returns a connection handle to the database. The name is in a
	// driver-specific format.
	NewClient(ctx context.Context, name string) (Client, error)
}

// ServerInfo represents the response a server gives witha GET request to '/'.
type ServerInfo interface {
	// Response is the full response, unparsed.
	Response() json.RawMessage
	// Version should return the version string from the top level of the response.
	Version() string
	// Vendor should return the name of the vendor.
	Vendor() string
	// VendorVersion should return the vendor version number.
	VendorVersion() string
}

// Client is a connection to a database server.
type Client interface {
	// ServerInfo returns the server implementation's details.
	ServerInfo(ctx context.Context, options map[string]interface{}) (ServerInfo, error)
	AllDBs(ctx context.Context, options map[string]interface{}) ([]string, error)
	// DBExists returns true if the database exists.
	DBExists(ctx context.Context, dbName string, options map[string]interface{}) (bool, error)
	// CreateDB creates the requested DB. The dbName is validated as a valid
	// CouchDB database name prior to calling this function, so the driver can
	// assume a valid name.
	CreateDB(ctx context.Context, dbName string, options map[string]interface{}) error
	// DestroyDB deletes the requested DB.
	DestroyDB(ctx context.Context, dbName string, options map[string]interface{}) error
	// DB returns a handleto the requested database
	DB(ctx context.Context, dbName string, options map[string]interface{}) (DB, error)
}

// Authenticator is an optional interface that may be implemented by a Client
// that supports authenitcated connections.
type Authenticator interface {
	// Authenticate attempts to authenticate the client using an authenticator.
	// If the authenticator is not known to the client, an error should be
	// returned.
	Authenticate(ctx context.Context, authenticator interface{}) error
}

// DBInfo provides statistics about a database.
type DBInfo struct {
	Name           string `json:"db_name"`
	CompactRunning bool   `json:"compact_running"`
	DocCount       int64  `json:"doc_count"`
	DeletedCount   int64  `json:"doc_del_count"`
	UpdateSeq      string `json:"update_seq"`
	DiskSize       int64  `json:"disk_size"`
	ActiveSize     int64  `json:"data_size"`
	ExternalSize   int64  `json:"-"`
}

// Members represents the members of a database security document.
type Members struct {
	Names []string `json:"names,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// Security represents a database security document.
type Security struct {
	Admins  Members `json:"admins"`
	Members Members `json:"members"`
}

// DB is a database handle.
type DB interface {
	// AllDocs returns all of the documents in the database, subject to the
	// options provided.
	AllDocs(ctx context.Context, options map[string]interface{}) (Rows, error)
	// Get fetches the requested document from the database, and unmarshals it
	// into doc.
	Get(ctx context.Context, docID string, doc interface{}, options map[string]interface{}) error
	// CreateDoc creates a new doc, with a server-generated ID.
	CreateDoc(ctx context.Context, doc interface{}) (docID, rev string, err error)
	// Put writes the document in the database.
	Put(ctx context.Context, docID string, doc interface{}) (rev string, err error)
	// Delete marks the specified document as deleted.
	Delete(ctx context.Context, docID, rev string) (newRev string, err error)
	// Info returns information about the database
	Info(ctx context.Context) (*DBInfo, error)
	// Compact initiates compaction of the database.
	Compact(ctx context.Context) error
	// CompactView initiates compaction of the view.
	CompactView(ctx context.Context, ddocID string) error
	// ViewCleanup cleans up stale view files.
	ViewCleanup(ctx context.Context) error
	// Security returns the database's security document.
	Security(ctx context.Context) (*Security, error)
	// SetSecurity sets the database's security document.
	SetSecurity(ctx context.Context, security *Security) error
	// Changes returns a Rows iterator for the changes feed. In continuous mode,
	// the iterator will continue indefinately, until Close is called.
	Changes(ctx context.Context, options map[string]interface{}) (Changes, error)
	// BulkDocs alls bulk create, update and/or delete operations. It returns an
	// iterator over the results.
	BulkDocs(ctx context.Context, docs ...interface{}) (BulkResults, error)
	// PutAttachment uploads an attachment to the specified document, returning
	// the new revision.
	PutAttachment(ctx context.Context, docID, rev, filename, contentType string, body io.Reader) (newRev string, err error)
	// GetAttachment fetches an attachment for the associated document ID. rev
	// may be an empty string to fetch the most recent document version.
	GetAttachment(ctx context.Context, docID, rev, filename string) (contentType string, md5sum Checksum, body io.ReadCloser, err error)
	// DeleteAttachment deletes an attachment from a document, returning the
	// document's new revision.
	DeleteAttachment(ctx context.Context, docID, rev, filename string) (newRev string, err error)
	// Query performs a query against a view, subject to the options provided.
	// ddoc will be the design doc name without the '_design/' previx.
	// view will be the view name without the '_view/' prefix.
	Query(ctx context.Context, ddoc, view string, options map[string]interface{}) (Rows, error)
}

// Finder is an optional interface which may be implemented by a database. The
// Finder interface provides access to the new (in CouchDB 2.0) MongoDB-style
// query interface.
type Finder interface {
	// Find executes a query using the new /_find interface. If query is a
	// string, []byte, or json.RawMessage, it should be treated as a raw JSON
	// payload. Any other type should be marshaled to JSON.
	Find(ctx context.Context, query interface{}) (Rows, error)
	// CreateIndex creates an index if it doesn't already exist. If the index
	// already exists, it should do nothing. ddoc and name may be empty, in
	// which case they should be provided by the backend. If index is a string,
	// []byte, or json.RawMessage, it should be treated as a raw JSON payload.
	// Any other type should be marshaled to JSON.
	CreateIndex(ctx context.Context, ddoc, name string, index interface{}) error
	// GetIndexes returns a list of all indexes in the database.
	GetIndexes(ctx context.Context) ([]Index, error)
	// Delete deletes the requested index.
	DeleteIndex(ctx context.Context, ddoc, name string) error
}

// Index is a MonboDB-style index definition.
type Index struct {
	DesignDoc  string      `json:"ddoc,omitempty"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Definition interface{} `json:"def"`
}

// Checksum is a 128-bit MD5 checksum of a file's content.
type Checksum [16]byte

// AttachmentMetaer is an optional interface which may be satisfied by a
// DB. If satisfied, it may be used to fetch meta data about an attachment. If
// not satisfied, GetAttachment will be used instead.
type AttachmentMetaer interface {
	// GetAttachmentMeta returns meta information about an attachment.
	GetAttachmentMeta(ctx context.Context, docID, rev, filename string) (contentType string, md5sum Checksum, err error)
}

// BulkResult is the result of a single doc update in a BulkDocs request.
type BulkResult struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Error error
}

// BulkResults is an iterator over the results for a BulkDocs call.
type BulkResults interface {
	// Next is called to populate *BulkResult with the values of the next bulk
	// result in the set.
	//
	// Next should return io.EOF when there are no more results.
	Next(*BulkResult) error
	// Close closes the bulk results iterator.
	Close() error
}

// Rever is an optional interface that may be implemented by a database. If not
// implemented by the driver, the Get method will be used to emulate the
// functionality.
type Rever interface {
	// Rev returns the most current revision of the requested document.
	Rev(ctx context.Context, docID string) (rev string, err error)
}

// DBFlusher is an optional interface that may be implemented by a database
// that can force a flush of the database backend file(s) to disk or other
// permanent storage.
type DBFlusher interface {
	// Flush requests a flush of disk cache to disk or other permanent storage.
	//
	// See http://docs.couchdb.org/en/2.0.0/api/database/compact.html#db-ensure-full-commit
	Flush(ctx context.Context) error
}

// Copier is an optional interface that may be implemented by a DB.
//
// If a DB does implement Copier, Copy() functions will use it.  If a DB does
// not implement the Copier interface, or if a call to Copy() returns an
// http.StatusUnimplemented, the driver will emulate a copy by doing
// a GET followed by PUT.
type Copier interface {
	Copy(ctx context.Context, targetID, sourceID string, options map[string]interface{}) (targetRev string, err error)
}

// Configer is an optional interface that may be implemented by a Client.
//
// If a Client does implement Configer, it allows backend configuration
// to be queried and modified via the API.
type Configer interface {
	Config(ctx context.Context) (Config, error)
}

// Config is the minimal interface that a Config backend must implement.
type Config interface {
	GetAll(ctx context.Context) (config map[string]map[string]string, err error)
	Set(ctx context.Context, secName, key, value string) error
	Delete(ctx context.Context, secName, key string) error
}

// ConfigSection is an optional interface that may be implemented by a Config
// backend. If not implemented, it will be emulated with GetAll() and SetAll().
// The only reason for a config backend to implement this interface is if
// reading a config section alone can be more efficient than reading the entire
// configuration for the specific storage backend.
type ConfigSection interface {
	GetSection(ctx context.Context, secName string) (section map[string]string, err error)
}

// ConfigItem is an optional interface that may be implemented by a Config
// backend. If not implemented, it will be emulated with GetAll() and SetAll().
// The only reason for a config backend to implement this interface is if
// reading a single config value alone can be more efficient than reading the
// entire configuration for the specific storage backend.
type ConfigItem interface {
	Get(ctx context.Context, secName, key string) (value string, err error)
}
