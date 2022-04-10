package record

/*
	Some basic thoughts about the record provider:
		* Should try to first follow URI references in records, followed by a prioritized list of
		  clients in a config file.
*/

type Attestation interface {
}

type Body interface {
}

// TODO: This shoould probably just use something like OCI Content descriptors
type Material struct {
	MediaType string
	Digest    string
	Uri       string
}

type Record interface {
	GetType() string
	GetMaterials() []Material
	GetData() interface{}
}

type ManagerOpts struct {
	IsTest bool
}

// TODO: Should probably create some creator functions for this
type Manager struct {
	Opts ManagerOpts
	// TODO: There should be a way to create a way to prioritize different backends for different types of records
	Clients []RecordClient
}

// TODO: This client is intended to follow URIs found in records. This will require some consensus in the community
type URIFollowerClient struct{}

type RecordClient interface {
	GetRecord(hash string) (Record, error)
}

func (m *Manager) GetRecord(hash string) (Record, error) {
	// TODO: Figure out if we fetch all records in all places or just the first record found
	for _, c := range m.Clients {
		r, err := c.GetRecord(hash)
		if err != nil {
			return nil, err
		}
		// Record found so return it
		if r != nil {
			return r, nil
		}
		// Record not found so go to the next client
	}
	// Record not found
	return nil, nil
}
