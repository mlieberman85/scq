package graph

import (
	"github.com/mlieberman85/scq/pkg/record"
)

type SupplyChainGraph struct {
	Nodes         map[string]*record.Record
	Edges         map[string]map[string]struct{} // Should this be a set of some sort?
	RecordManager *record.Manager
}

func (scg *SupplyChainGraph) GenerateFromHash(hash string) error {
	if _, ok := scg.Edges[hash]; !ok {
		scg.Edges[hash] = make(map[string]struct{})
	}

	r, err := scg.RecordManager.GetRecord(hash)
	if _, ok := scg.Nodes[hash]; !ok {
		// TODO: How do we want to handle mutliple data all associated with a single hash?
		scg.Nodes[hash] = &r
	}
	if err != nil {
		return err
	}

	for _, m := range r.GetMaterials() {
		// Check if seen hash before in these edges.
		if _, ok := scg.Edges[hash][m.Digest]; !ok {
			scg.Edges[hash][m.Digest] = struct{}{}
			// Check if processed hash before
			if _, node := scg.Nodes[m.Digest]; !node {
				err = scg.GenerateFromHash(m.Digest)
				if err != nil {
					// TODO: Ignore only errors where no matching entries are found
				}
			}
		}
	}

	return nil
}
