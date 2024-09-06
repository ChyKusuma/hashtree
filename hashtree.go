package hashtree

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

// HashTreeNode represents a node in the hash tree
type HashTreeNode struct {
	Hash  []byte        `json:"hash"`
	Left  *HashTreeNode `json:"left,omitempty"`
	Right *HashTreeNode `json:"right,omitempty"`
}

// Compute the hash of a given data slice
func computeHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// BuildHashTree builds the hash tree from leaf nodes
func BuildHashTree(leaves [][]byte) *HashTreeNode {
	nodes := make([]*HashTreeNode, len(leaves))
	for i, leaf := range leaves {
		nodes[i] = &HashTreeNode{Hash: computeHash(leaf)}
	}

	for len(nodes) > 1 {
		var nextLevel []*HashTreeNode
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				left, right := nodes[i], nodes[i+1]
				hash := computeHash(append(left.Hash, right.Hash...))
				nextLevel = append(nextLevel, &HashTreeNode{Hash: hash, Left: left, Right: right})
			} else {
				nextLevel = append(nextLevel, nodes[i])
			}
		}
		nodes = nextLevel
	}

	return nodes[0]
}

// Generate random data of specified length
func generateRandomData(size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Save root hash to file
func saveRootHashToFile(root *HashTreeNode, filename string) error {
	return ioutil.WriteFile(filename, root.Hash, 0644)
}

// Load root hash from file
func loadRootHashFromFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// Save leaf node data to LevelDB
func saveLeavesToDB(db *leveldb.DB, leaves [][]byte) error {
	for i, leaf := range leaves {
		key := fmt.Sprintf("leaf-%d", i)
		err := db.Put([]byte(key), leaf, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fetch leaf from LevelDB
func fetchLeafFromDB(db *leveldb.DB, key string) ([]byte, error) {
	return db.Get([]byte(key), nil)
}

// Print the root hash of the hash tree
func printRootHash(root *HashTreeNode) {
	fmt.Printf("Root Hash: %x\n", root.Hash)
}
