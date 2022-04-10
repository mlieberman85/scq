package graph

import (
	"github.com/mlieberman85/scq/pkg/record"
)

type SupplyChainGraph struct {
	Nodes         map[string]*record.Record
	Edges         map[string][]string
	RecordManager *record.Manager
}

func (scg *SupplyChainGraph) GenerateFromHash(hash string) error {
	if _, ok := scg.Edges[hash]; !ok {
		scg.Edges[hash] = make([]string, 0)
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
		scg.Edges[hash] = append(scg.Edges[hash], m.Digest)
		err = scg.GenerateFromHash(m.Digest)
		if err != nil {
			// TODO: Ignore only errors where no matching entries are found
		}
	}

	return nil
}
