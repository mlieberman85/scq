# scq

This is a Supply Chain Query tool intended to query datastores containing attestations, SBOMs, and other supply chain metadata and build a graph that can be queried.

This is currently a POC and is being tested by storing attestations in mongodb and thus relies on mongo db for testing.

Right now the way you would test it out is:

```
go build
./scq test testdata/
cat testdata/foo.json | jq '.subject[0].digest.sha256' | xargs -I{} ./scq graph --db mongo --hash {} | jq | less
```

The above commands will store the testdata into mongodb and then generate a graph based on the hash from the `foo.json` test attestation. It will recursively query the mongodb until it can't find any attestations to follow.