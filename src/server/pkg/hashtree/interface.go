package hashtree

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ErrCode identifies different kinds of errors returned by methods in
// HashTree below. The ErrCode of any such error can be retrieved with Code().
type ErrCode uint8

const (
	// OK is returned on success
	OK ErrCode = iota

	// Unknown is returned by Code() when an error wasn't emitted by the HashTree
	// implementation.
	Unknown

	// Internal is returned when a HashTree encounters a bug (usually due to the
	// violation of an internal invariant).
	Internal

	// CannotDeserialize is returned when Unmarshal(bytes) fails, perhaps due to
	// 'bytes' being corrupted.
	CannotDeserialize

	// Unsupported is returned when Unmarshal(bytes) encounters an unsupported
	// (likely old) serialized HashTree.
	Unsupported

	// PathNotFound is returned when Get() or DeleteFile() is called with a path
	// that doesn't lead to a node.
	PathNotFound

	// MalformedGlob is returned when Glob() is called with an invalid glob
	// pattern.
	MalformedGlob

	// PathConflict is returned when a path that is expected to point to a
	// directory in fact points to a file, or the reverse. For example:
	// 1. PutFile is called with a path that points to a directory.
	// 2. PutFile is called with a path that contains a prefix that
	//    points to a file.
	// 3. Merge is forced to merge a directory into a file
	PathConflict
)

// HashTree is the signature of a hash tree provided by this library. To get a
// new hash tree, see hashtree.NewHashTree().
type HashTree interface {
	// Open converts this to an OpenHashTree, which can be modified.
	Open() OpenHashTree

	// Get retrieves the contents of a regular file.
	Get(path string) (*NodeProto, error)

	// List retrieves the list of files and subdirectories of the directory at
	// 'path'.
	List(path string) ([]*NodeProto, error)

	// Glob returns a list of files and directories that match 'pattern'.
	Glob(pattern string) ([]*NodeProto, error)

	// Marshal serializes a HashTree so that it can be persisted. Also see
	// Unmarshal().
	Marshal() ([]byte, error)
}

// OpenNode is similar to NodeProto, except that it doesn't include the Hash or
// Size fields (which may be inaccurate in an OpenHashTree)
type OpenNode struct {
	Name string

	FileNode *FileNodeProto
	DirNode  *DirectoryNodeProto
}

// OpenHashTree is like HashTree, except that it can be modified. Once an
// OpenHashTree is Finish()ed, the hash and size stored with each node will be
// updated (until then, the hashes and sizes stored in an OpenHashTree will be
// stale, which is why this interface has no functions for reading data in it).
// We separated HashTree and OpenHashTree, instead of re-hashing a HashTree
// after each operation for performance; re-hashing can consume a lot of time
// if e.g. you need to PutFile 100K files in a row and re-hash after each one.
type OpenHashTree interface {
	// Get retrieves the contents of a regular file.
	GetOpen(path string) (*OpenNode, error)

	// PutFile appends data to a file (and creates the file if it doesn't exist).
	// This invalidates the hashes in the tree. Calling this marks the HashTree
	// unfinished (see Finish() and IsFinished()).
	PutFile(path string, blockRefs []*pfs.BlockRef) error

	// PutDir creates a directory (or does nothing if one exists). Calling this
	// marks the HashTree unfinished (see Finish() and IsFinished()).
	PutDir(path string) error

	// DeleteFile deletes a regular file or directory (along with its children).
	// Calling this marks the HashTree unfinished (see Finish() and
	// IsFinished()).
	DeleteFile(path string) error

	// Merge adds all of the files and directories in each tree in 'trees' into
	// this tree.
	Merge(trees []OpenHashTree) error

	// Finish updates all of the hashes and sizes of the nodes in the HashTree.
	Finish() (HashTree, error)
}
