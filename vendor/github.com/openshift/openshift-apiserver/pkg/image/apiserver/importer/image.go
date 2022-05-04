package importer

import (
	"encoding/json"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/registry/api/errcode"
	godigest "github.com/opencontainers/go-digest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/api/image/dockerpre012"

	imageapi "github.com/openshift/openshift-apiserver/pkg/image/apis/image"
	imagedockerpre012 "github.com/openshift/openshift-apiserver/pkg/image/apis/image/dockerpre012"
)

func schema1ToImage(manifest *schema1.SignedManifest, d godigest.Digest) (*imageapi.Image, error) {
	if len(manifest.History) == 0 {
		return nil, fmt.Errorf("image has no v1Compatibility history and cannot be used")
	}
	dockerImage, err := unmarshalDockerImage([]byte(manifest.History[0].V1Compatibility))
	if err != nil {
		return nil, err
	}
	mediatype, payload, err := manifest.Payload()
	if err != nil {
		return nil, err
	}

	if len(manifest.Canonical) == 0 {
		return nil, fmt.Errorf("unable to load canonical representation from schema1 manifest")
	}
	payloadDigest := godigest.FromBytes(manifest.Canonical)
	if len(d) > 0 && payloadDigest != d {
		return nil, fmt.Errorf("content integrity error: the schema 1 manifest retrieved with digest %s does not match the digest calculated from the content %s", d, payloadDigest)
	}
	dockerImage.ID = payloadDigest.String()

	image := &imageapi.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: dockerImage.ID,
		},
		DockerImageMetadata:          *dockerImage,
		DockerImageManifest:          string(payload),
		DockerImageManifestMediaType: mediatype,
		DockerImageMetadataVersion:   "1.0",
	}

	return image, nil
}

// schema2OrOCIToImage converts a docker schema 2 or an oci schema manifest into an Image.
func schema2OrOCIToImage(manifest distribution.Manifest, imageConfig []byte, d godigest.Digest) (*imageapi.Image, error) {
	mediatype, payload, err := manifest.Payload()
	if err != nil {
		return nil, err
	}

	dockerImage, err := unmarshalDockerImage(imageConfig)
	if err != nil {
		return nil, err
	}

	payloadDigest := godigest.FromBytes(payload)
	if len(d) > 0 && payloadDigest != d {
		return nil, fmt.Errorf(
			"content integrity error: the manifest retrieved (media type: %s) "+
				"with digest %s does not match the digest calculated from "+
				"the content %s",
			mediatype,
			d,
			payloadDigest,
		)
	}
	dockerImage.ID = payloadDigest.String()

	image := &imageapi.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: dockerImage.ID,
		},
		DockerImageMetadata:          *dockerImage,
		DockerImageManifest:          string(payload),
		DockerImageConfig:            string(imageConfig),
		DockerImageManifestMediaType: mediatype,
		DockerImageMetadataVersion:   "1.0",
	}

	return image, nil
}

func unmarshalDockerImage(body []byte) (*imageapi.DockerImage, error) {
	var image dockerpre012.DockerImage
	if err := json.Unmarshal(body, &image); err != nil {
		return nil, err
	}
	dockerImage := &imageapi.DockerImage{}
	if err := imagedockerpre012.Convert_dockerpre012_DockerImage_To_image_DockerImage(&image, dockerImage, nil); err != nil {
		return nil, err
	}
	return dockerImage, nil
}

func isDockerError(err error, code errcode.ErrorCode) bool {
	switch t := err.(type) {
	case errcode.Errors:
		for _, err := range t {
			if isDockerError(err, code) {
				return true
			}
		}
	case errcode.ErrorCode:
		if code == t {
			return true
		}
	case errcode.Error:
		if t.ErrorCode() == code {
			return true
		}
	}
	return false
}
