package record

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/in-toto/in-toto-golang/in_toto"
	rekor_client "github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/index"
	"github.com/sigstore/rekor/pkg/generated/models"
)

// TODO: This needs to be rewritten. A lot of this is more or less copied from various sigstore tools
// since the Rekor API is mostly generated openapi code.

const rekor_server = "https://rekor.sigstore.dev"

type RekorRecord struct {
	Body        Body
	Attestation Attestation
}

type RekorClient struct {
	client *rekor_client.Rekor
}

func (rc *RekorClient) GetRecord(hash string) (Record, error) {
	UUIDs, err := searchRekor(rc.client, hash)
	if err != nil {
		return nil, err
	}

	for _, u := range UUIDs {
		entry, err := getEntry(rc.client, u)
		if err != nil {
			log.Fatal(err)
		}
		record, err := parseEntry(entry)
		if err != nil {
			log.Fatal(err)
		}
		return record, nil
	}
	return nil, fmt.Errorf("Unreachable!")
}

func (r *RekorRecord) GetData() interface{} {
	return r
}

func (r *RekorRecord) GetType() string {
	return "rekor"
}

func (r *RekorRecord) GetMaterials() []Material {
	var ms []Material
	for _, predicateMaterial := range r.Attestation.(*in_toto.ProvenanceStatement).Predicate.Materials {
		// For now just get the first value in the map.
		var d string
		for _, digest := range predicateMaterial.Digest {
			d = digest
			break
		}
		m := Material{
			MediaType: "todo", // TODO: Get folks to use OCI in their attestations
			Digest:    d,
		}
		ms = append(ms, m)
	}

	return ms
}

func parseStatement(p string) (*in_toto.ProvenanceStatement, error) {
	ps := in_toto.ProvenanceStatement{}
	payload, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, &ps); err != nil {
		return nil, err
	}
	return &ps, nil
}

func searchRekor(rekorClient *rekor_client.Rekor, sha string) ([]string, error) {
	UUIDs := make(map[string]struct{})

	searchIndexParams := index.NewSearchIndexParams()
	searchIndexParams.SetTimeout(time.Minute) // TODO: Make configurable
	searchIndexParams.Query = &models.SearchIndex{}

	searchIndexParams.Query.Hash = sha

	resp, err := rekorClient.Index.SearchIndex(searchIndexParams)

	if err != nil {
		return nil, err
	}

	if len(resp.Payload) == 0 {
		return nil, fmt.Errorf("no matching entries found")
	}

	for _, v := range resp.GetPayload() {
		UUIDs[v] = struct{}{}
	}

	keys := make([]string, 0, len(UUIDs))
	for k := range UUIDs {
		keys = append(keys, k)
	}

	return keys, nil
}

func getEntry(rekorClient *rekor_client.Rekor, UUID string) (*models.LogEntryAnon, error) {
	getLogEntryByUUIDParams := entries.NewGetLogEntryByUUIDParams()
	getLogEntryByUUIDParams.SetTimeout(time.Minute)
	getLogEntryByUUIDParams.EntryUUID = UUID
	response, err := rekorClient.Entries.GetLogEntryByUUID(getLogEntryByUUIDParams)
	if err != nil {
		return nil, err
	}

	// NOTE: Because of code gen, the response payload is just a list of length 1
	// except in cases of error so we can just return the first thing.
	for _, entry := range response.Payload {
		return &entry, nil
	}
	return nil, fmt.Errorf("Payload was empty")
}

func parseEntry(entry *models.LogEntryAnon) (*RekorRecord, error) {
	b, err := base64.StdEncoding.DecodeString(entry.Body.(string))
	if err != nil {
		return nil, err
	}

	pe, err := models.UnmarshalProposedEntry(bytes.NewReader(b), runtime.JSONConsumer())
	if err != nil {
		return nil, err
	}

	switch pe.Kind() {
	case "intoto":
		bodyJson, err := pe.(*models.Intoto).MarshalJSON()
		var intoto models.Intoto
		err = json.Unmarshal(bodyJson, &intoto)
		if err != nil {
			return nil, err
		}

		statement, err := parseStatement(string(entry.Attestation.Data))
		if err != nil {
			return nil, err
		}

		return &RekorRecord{
			Body:        string(bodyJson),
			Attestation: statement,
		}, nil

	default:
		return nil, fmt.Errorf("%s is an unsupported entry type", pe.Kind())
	}
}
