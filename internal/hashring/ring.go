package hashring

import (
	"hash/crc32"
	"sort"
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

func (r *Ring) AddNode(nodeAddr string) {
	hashed := crc32.ChecksumIEEE([]byte(nodeAddr))
	r.nodes = append(r.nodes, hashed)
	sort.Slice(r.nodes, func(i, j int) bool { return int(r.nodes[i]) < int(r.nodes[j]) })
	r.nodeMap[hashed] = nodeAddr
}

func (r *Ring) GetNode(key string) string {
	hashed := crc32.ChecksumIEEE([]byte(key))
	i := sort.Search(len(r.nodes), func(i int) bool { return r.nodes[i] >= hashed })
	if i == len(r.nodes) {
		return r.nodeMap[r.nodes[0]]
	}
	return r.nodeMap[r.nodes[i]]
}
