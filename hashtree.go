package hashtree

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"syscall"

	"github.com/syndtr/goleveldb/leveldb"
)

const maxFileSize = 1 << 30 // 1 GiB max file size for memory mapping

// HashTreeNode represents a node in the hash tree
type HashTreeNode struct {
	Hash  []byte        `json:"hash"`            // Hash of the node's data
	Left  *HashTreeNode `json:"left,omitempty"`  // Left child node
	Right *HashTreeNode `json:"right,omitempty"` // Right child node
}

// Compute the hash of a given data slice
func computeHash(data []byte) []byte {
	hash := sha256.Sum256(data) // Compute SHA-256 hash
	return hash[:]
}

// BuildHashTree builds the hash tree from leaf nodes
func BuildHashTree(leaves [][]byte) *HashTreeNode {
	// Create leaf nodes
	nodes := make([]*HashTreeNode, len(leaves))
	for i, leaf := range leaves {
		nodes[i] = &HashTreeNode{Hash: computeHash(leaf)}
	}

	// Build the hash tree
	for len(nodes) > 1 {
		var nextLevel []*HashTreeNode
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				left, right := nodes[i], nodes[i+1]
				hash := computeHash(append(left.Hash, right.Hash...)) // Combine and hash
				nextLevel = append(nextLevel, &HashTreeNode{Hash: hash, Left: left, Right: right})
			} else {
				nextLevel = append(nextLevel, nodes[i]) // Handle odd number of nodes
			}
		}
		nodes = nextLevel // Move up a level
	}

	return nodes[0] // Return the root of the hash tree
}

// Generate random data of specified length
func GenerateRandomData(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data) // Fill with random data
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Save root hash to file
func SaveRootHashToFile(root *HashTreeNode, filename string) error {
	return ioutil.WriteFile(filename, root.Hash, 0644) // Save root hash to file
}

// Load root hash from file
func LoadRootHashFromFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename) // Read root hash from file
}

// Save leaf node data to LevelDB
func SaveLeavesToDB(db *leveldb.DB, leaves [][]byte) error {
	for i, leaf := range leaves {
		key := fmt.Sprintf("leaf-%d", i)
		err := db.Put([]byte(key), leaf, nil) // Store leaf node in LevelDB
		if err != nil {
			return err
		}
	}
	return nil
}

// Fetch leaf from LevelDB
func FetchLeafFromDB(db *leveldb.DB, key string) ([]byte, error) {
	return db.Get([]byte(key), nil) // Retrieve leaf node from LevelDB
}

// Print the root hash of the hash tree
func PrintRootHash(root *HashTreeNode) {
	fmt.Printf("Root Hash: %x\n", root.Hash) // Print root hash
}

// Batch operations for LevelDB to improve performance
func SaveLeavesBatchToDB(db *leveldb.DB, leaves [][]byte) error {
	batch := new(leveldb.Batch)
	for i, leaf := range leaves {
		key := fmt.Sprintf("leaf-%d", i)
		batch.Put([]byte(key), leaf) // Add leaf node to batch
	}
	return db.Write(batch, nil) // Execute batch write
}

// Handle concurrent access to LevelDB (basic example)
func FetchLeafConcurrent(db *leveldb.DB, key string) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *leveldb.Transaction) error {
		var err error
		value, err = txn.Get([]byte(key), nil) // Read from LevelDB transaction
		return err
	})
	return value, err
}

// Define a suitable maxFileSize based on your needs and system constraints
const maxFileSize = 1 << 30 // 1 GiB max file size for memory mapping

// Update maxFileSize based on specific needs
func setMaxFileSize(sizeInGiB int) {
	// Ensure size is reasonable and within system limits
	if sizeInGiB <= 0 {
		fmt.Println("Invalid size. Must be greater than 0.")
		return
	}
	maxFileSize = sizeInGiB * (1 << 30) // Convert GiB to bytes
}

// MemoryMapFile maps a file into memory with size checks
func MemoryMapFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}

	size := stat.Size()
	if size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %d bytes", maxFileSize)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("error mapping file: %w", err)
	}

	return data, nil
}

// UnmapFile unmaps a file from memory with error handling
func UnmapFile(data []byte) error {
	if err := syscall.Munmap(data); err != nil {
		return fmt.Errorf("error unmapping file: %w", err)
	}
	return nil
}

// Concurrency control for memory-mapped file access
var mu sync.Mutex

func SafeMemoryMapFile(filename string) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return MemoryMapFile(filename)
}

func SafeUnmapFile(data []byte) error {
	mu.Lock()
	defer mu.Unlock()
	return UnmapFile(data)
}
