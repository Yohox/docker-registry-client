package registry

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

func (r *Registry) Manifest(repository, reference string) (*schema1.SignedManifest, error) {
	url := r.generateUrl("/v2/%s/manifests/%s", repository, reference)
	r.logf("registry.manifest.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", schema1.MediaTypeManifest)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	signedManifest := &schema1.SignedManifest{}
	err = signedManifest.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}

	return signedManifest, nil
}

func (r *Registry) ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error) {
	url := r.generateUrl("/v2/%s/manifests/%s", repository, reference)
	r.logf("registry.manifest.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", schema2.MediaTypeManifest)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	deserialized := &schema2.DeserializedManifest{}
	err = deserialized.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return deserialized, nil
}

func (r *Registry) ManifestDigest(repository, reference string) (digest.Digest, error) {
	url := r.generateUrl("/v2/%s/manifests/%s", repository, reference)
	r.logf("registry.manifest.head url=%s repository=%s reference=%s", url, repository, reference)

	resp, err := r.client.Head(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return digest.Parse(resp.Header.Get("Docker-Content-Digest"))
}

func (r *Registry) DeleteManifest(repository string, digest digest.Digest) error {
	url := r.generateUrl("/v2/%s/manifests/%s", repository, digest)
	r.logf("registry.manifest.delete url=%s repository=%s reference=%s", url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Registry) PutManifest(repository, reference string, manifest distribution.Manifest) error {
	url := r.generateUrl("/v2/%s/manifests/%s", repository, reference)
	r.logf("registry.manifest.put url=%s repository=%s reference=%s", url, repository, reference)

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(payload)
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mediaType)
	resp, err := r.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}
