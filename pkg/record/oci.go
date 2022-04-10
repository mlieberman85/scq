package record

import (
	"context"
	"encoding/json"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/sigstore/cosign/pkg/cosign"
)

// WIP

// Uses cosign to do most of the heavy lifting.
type OCIRecord struct {
	Attestation in_toto.ProvenanceStatement
}

func (o *OCIRecord) GetType() string {
	return "slsa"
}

func (o *OCIRecord) GetData() interface{} {
	return o.Attestation
}

func (o *OCIRecord) GetMaterials() []Material {
	// TODO: Make wrapper for in_toto predicate types as that's going to be a common one
	var ms []Material
	for _, predicateMaterial := range o.Attestation.Predicate.Materials {
		// For now just get the first hash in the map.
		var d string
		for _, digest := range predicateMaterial.Digest {
			d = digest
			break
		}

		m := Material{
			MediaType: "todo",
			Digest:    d,
			Uri:       predicateMaterial.URI,
		}
		ms = append(ms, m)
	}

	return ms
}

// TODO: Set up trusted keys for records
// This for now is intended to purely represent records related to a particular image
type OCIClient struct {
	//trustedKeyRefs []string
	imageURI string
}

func (o *OCIClient) GetRecord(hash string) (Record, error) {
	ctx := context.Background()

	d, err := name.NewDigest(o.imageURI)
	if err != nil {
		return nil, err
	}

	as, err := cosign.FetchAttestationsForReference(ctx, d)
	if err != nil {
		return nil, err
	}

	// TODO: What to do when multiple attestations are found?
	if len(as) == 0 {
		return nil, nil
	}

	var provenance in_toto.ProvenanceStatement
	// TODO: This currently doesn't handle anything but slsa provenance
	err = json.Unmarshal([]byte(as[0].PayLoad), &provenance)
	if err != nil {
		return nil, err
	}

	r := OCIRecord{
		Attestation: provenance,
	}
	return &r, nil
}
