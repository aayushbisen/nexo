package hashring

import (
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
)

type Ring struct {
	nodes   []uint32
	nodeMap map[uint32]string
}

func New() *Ring {
	nodeM := make(map[uint32]string)
	nodess := []uint32{}
	return &Ring{
		nodeMap: nodeM,
		nodes:   nodess,
	}
}

func (r *Ring) AddNode(nodeAddr string, replicas int) {
	if replicas <= 0 {
		panic("hashring: replicas must be > 0")
	}
	for u := range replicas {
		hashed := crc32.ChecksumIEEE([]byte(nodeAddr + "$" + strconv.Itoa(u)))
		r.nodes = append(r.nodes, hashed)
		r.nodeMap[hashed] = nodeAddr
	}
	slices.Sort(r.nodes)
}

func (r *Ring) GetNode(key string) string {
	hashed := crc32.ChecksumIEEE([]byte(key))
	i := sort.Search(len(r.nodes), func(i int) bool { return r.nodes[i] >= hashed })
	if i == len(r.nodes) {
		return r.nodeMap[r.nodes[0]]
	}
	return r.nodeMap[r.nodes[i]]
}
