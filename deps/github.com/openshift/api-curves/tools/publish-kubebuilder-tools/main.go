package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"gopkg.in/yaml.v3"
)

func main() {
	pullSecretFile := flag.String("pull-secret", "", "The pull secret to use for the kubebuilder tools")
	version := flag.String("version", "", "The version of the kubebuilder tools to publish. This should be a Kubernetes version that the build is based upon.")
	outputDir := flag.String("output-dir", "", "The output directory to write the kubebuilder tools to")
	payload := flag.String("payload", "", "The payload to use for building the kubebuilder tools archives. This should be in the format registry.ci.openshift.org/ocp/release:<version>")
	skipUpload := flag.Bool("skip-upload", false, "Skip uploading the artifacts created to the openshift-gce-devel/openshift-kubebuilder-tools bucket")
	indexFile := flag.String("index-file", "envtest-releases.yaml", "The index file to use for the kubebuilder tools")

	flag.Parse()

	if *pullSecretFile == "" {
		panic("pull-secret is required")
	}

	if *version == "" {
		panic("version is required")
	}

	if *outputDir == "" {
		panic("output is required")
	}

	// We only expect Kubernetes versions that begin with the v prefix.
	if !strings.HasPrefix(*version, "v") {
		panic("Kubernetes version must begin with the v prefix. Example: v1.33.2")
	}

	// Decrypt the pull secret to get the bearer token.
	registryAuthToken, err := getRegistryAuthToken(*pullSecretFile)
	if err != nil {
		panic(err)
	}

	// We only expect images from the internal releases
	if *payload == "" || !strings.HasPrefix(*payload, "registry.ci.openshift.org/ocp/release:") {
		panic("payload is required and must be a valid payload starting with \"registry.ci.openshift.org/ocp/release:\"")
	}

	payloadVersion := strings.TrimPrefix(*payload, "registry.ci.openshift.org/ocp/release:")

	// Download the image-references and convert to a map of image name to digest
	manifests, err := getReleaseImages(payloadVersion, registryAuthToken)
	if err != nil {
		panic(err)
	}

	// Extract the kube-apiserver binaries from the installer-kube-apiserver-artifacts image
	if err := getKubeAPIServerBins(*outputDir, manifests, registryAuthToken); err != nil {
		panic(err)
	}

	// Extract the etcd binaries from the installer-etcd-artifacts image
	if err := getEtcdBins(*outputDir, manifests, registryAuthToken); err != nil {
		panic(err)
	}

	// Build the envtest archives for each os and arch combination
	if err := buildEnvtestTars(*outputDir, *version); err != nil {
		panic(err)
	}

	if *skipUpload {
		fmt.Printf("Archives written to %s\n", *outputDir)
		return
	}

	// Upload the tars created earlier to the public GCS bucket for general consumption
	if err := uploadArchives(*outputDir, *version); err != nil {
		panic(err)
	}

	// Update the index file with the new version
	if err := updateIndexFile(*outputDir, *version, *indexFile); err != nil {
		panic(err)
	}

	fmt.Printf("Archives uploaded to openshift-gce-devel/openshift-kubebuilder-tools for version %s\n", *version)
}

func getRegistryAuthToken(pullSecretFile string) (string, error) {
	pullSecretRaw, err := os.ReadFile(pullSecretFile)
	if err != nil {
		return "", err
	}

	var secret struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}

	if err := json.Unmarshal(pullSecretRaw, &secret); err != nil {
		return "", err
	}

	registryAuth, ok := secret.Auths["registry.ci.openshift.org"]
	if !ok {
		return "", errors.New("registry.ci.openshift.org not found in pull secret")
	}

	registryAuthToken, err := base64.StdEncoding.DecodeString(registryAuth.Auth)
	if err != nil {
		return "", err
	}

	if len(strings.Split(string(registryAuthToken), ":")) != 2 {
		return "", errors.New("password not found in pull secret")
	}

	return strings.Split(string(registryAuthToken), ":")[1], nil
}

func getReleaseImages(version string, registryToken string) (map[string]string, error) {
	releaseManifestRaw, err := downloadJSON(getRegistryURL("release", "manifests", version), registryToken)
	if err != nil {
		return nil, err
	}

	manifest := struct {
		FSLayers []struct {
			BlobSum string `json:"blobSum"`
		} `json:"fsLayers"`
	}{}

	if err := json.Unmarshal(releaseManifestRaw, &manifest); err != nil {
		return nil, err
	}

	if len(manifest.FSLayers) == 0 {
		return nil, errors.New("no fsLayers found in release manifest")
	}

	// The first fsLayer is the release image manifests, which has the image digests for the release images
	releaseImageLayer, close, err := downloadArchive(getRegistryURL("release", "blobs", manifest.FSLayers[0].BlobSum), registryToken)
	if err != nil {
		return nil, err
	}
	defer close()

	imageReferencesRaw, err := getFileFromArchive(releaseImageLayer, "release-manifests/image-references")
	if err != nil {
		return nil, err
	}

	imageReferences := struct {
		Spec struct {
			Tags []struct {
				Name string
				From struct {
					Name string `json:"name"`
				} `json:"from"`
			} `json:"tags"`
		} `json:"spec"`
	}{}

	if err := json.Unmarshal(imageReferencesRaw, &imageReferences); err != nil {
		return nil, err
	}

	images := map[string]string{}
	for _, tag := range imageReferences.Spec.Tags {
		images[tag.Name] = tag.From.Name
	}

	return images, nil
}

func getRegistryURL(image, kind, digest string) string {
	return fmt.Sprintf("https://registry.ci.openshift.org/v2/ocp/%s/%s/%s", image, kind, digest)
}

func downloadJSON(url string, registryToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", registryToken))

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// From an in-memory archive, traverse and find the file named and return it as a slice of bytes.
func getFileFromArchive(archive *tar.Reader, filename string) ([]byte, error) {
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if header.Name == filename {
			buf := bytes.NewBuffer(nil)
			_, err := io.Copy(buf, archive)
			if err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("file %s not found in archive", filename)
}

// Fetch a tar.gz (container image layer) into memory so that we can extract files from it.
func downloadArchive(url string, registryToken string) (*tar.Reader, func() error, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", registryToken))

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, err
	}

	close := resp.Body.Close

	if resp.StatusCode != http.StatusOK {
		close()
		return nil, nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	uncompressedStream, err := gzip.NewReader(resp.Body)
	if err != nil {
		close()
		return nil, nil, err
	}

	return tar.NewReader(uncompressedStream), close, nil
}

func getKubeAPIServerBins(dir string, manifests map[string]string, registryToken string) error {
	return getMultiArchBinariesFromImage(dir, manifests, registryToken, "installer-kube-apiserver-artifacts", "kube-apiserver")

}

func getEtcdBins(dir string, manifests map[string]string, registryToken string) error {
	return getMultiArchBinariesFromImage(dir, manifests, registryToken, "installer-etcd-artifacts", "etcd")
}

func getMultiArchBinariesFromImage(dir string, manifests map[string]string, registryToken string, imageName string, binaryName string) error {
	image, ok := manifests[imageName]
	if !ok {
		return fmt.Errorf("%q not found in release images", imageName)
	}

	imageStream, digest := getImageStreamAndDigest(image)

	imageManifestRaw, err := downloadJSON(getRegistryURL(imageStream, "manifests", digest), registryToken)
	if err != nil {
		return err
	}

	imageManifest := struct {
		Manifests []struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}{}

	if err := json.Unmarshal(imageManifestRaw, &imageManifest); err != nil {
		return err
	}

	if len(imageManifest.Manifests) != 1 {
		return fmt.Errorf("expected 1 image manifest for image stream %q, got 0 or more than 1", imageName)
	}

	imageLayerManifestRaw, err := downloadJSON(getRegistryURL(imageStream, "manifests", imageManifest.Manifests[0].Digest), registryToken)
	if err != nil {
		return err
	}

	imageLayerManifest := struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}{}

	if err := json.Unmarshal(imageLayerManifestRaw, &imageLayerManifest); err != nil {
		return err
	}

	// The last layer is the layer containing the binaries.
	imageLayer, close, err := downloadArchive(getRegistryURL(imageStream, "blobs", imageLayerManifest.Layers[len(imageLayerManifest.Layers)-1].Digest), registryToken)
	if err != nil {
		return err
	}
	defer close()

	for _, goos := range []string{"darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			binDir := filepath.Join(dir, goos, arch, "bin")
			if err := os.MkdirAll(binDir, 0755); err != nil {
				return err
			}

			binRaw, err := getFileFromArchive(imageLayer, fmt.Sprintf("usr/share/openshift/%s/%s/%s", goos, arch, binaryName))
			if err != nil {
				return err
			}

			if err := os.WriteFile(filepath.Join(binDir, binaryName), binRaw, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func getImageStreamAndDigest(image string) (string, string) {
	streamAndDigest := strings.TrimPrefix(image, "registry.ci.openshift.org/ocp/")
	parts := strings.Split(streamAndDigest, "@")
	return parts[0], parts[1]
}

func buildEnvtestTars(dir string, version string) error {
	for _, goos := range []string{"darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			out, err := os.Create(filepath.Join(dir, fmt.Sprintf("envtest-%s-%s-%s.tar.gz", version, goos, arch)))
			if err != nil {
				return err
			}
			defer out.Close()

			gzArchive := gzip.NewWriter(out)
			defer gzArchive.Close()

			binFS := os.DirFS(filepath.Join(dir, goos, arch))
			tarArchive := tar.NewWriter(gzArchive)
			defer tarArchive.Close()

			if err := addFS(tarArchive, binFS); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFS copied from std library for Tar, introduced in Go 1.22. Copy/paste for now.
func addFS(tw *tar.Writer, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		// TODO(#49580): Handle symlinks when fs.ReadLinkFS is available.
		if !info.Mode().IsRegular() {
			return errors.New("tar: cannot add non-regular file")
		}
		h, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		h.Name = name
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		f, err := fsys.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func uploadArchives(dir string, version string) error {
	gcsClient, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	for _, goos := range []string{"darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			archivePath := filepath.Join(dir, fmt.Sprintf("envtest-%s-%s-%s.tar.gz", version, goos, arch))
			if err := uploadArchive(gcsClient, archivePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func uploadArchive(gcsClient *storage.Client, archivePath string) error {
	gcsObj := gcsClient.Bucket("openshift-kubebuilder-tools").Object(filepath.Base(archivePath))
	gcsWriter := gcsObj.NewWriter(context.Background())
	defer gcsWriter.Close()

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(gcsWriter, f); err != nil {
		return err
	}

	return nil
}

type indexFile struct {
	Hash     string `yaml:"hash"`
	SelfLink string `yaml:"selfLink"`
}

// updateIndexFile adds the new version to the existing index file.
// The index file is used by the setup-envtest tool to find the download links for the envtest archives.
func updateIndexFile(dir, version, indexFileName string) error {
	indexFileRaw, err := os.ReadFile(indexFileName)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	index := struct {
		Releases map[string]map[string]indexFile `json:"releases"`
	}{}

	if indexFileRaw != nil {
		if err := yaml.Unmarshal(indexFileRaw, &index); err != nil {
			return fmt.Errorf("failed to unmarshal index file: %w", err)
		}
	} else {
		index.Releases = make(map[string]map[string]indexFile)
	}

	releaseIndexes := make(map[string]indexFile)
	for _, goos := range []string{"darwin", "linux"} {
		for _, arch := range []string{"amd64", "arm64"} {
			name := fmt.Sprintf("envtest-%s-%s-%s.tar.gz", version, goos, arch)

			archive, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				return fmt.Errorf("failed to open archive %s: %w", name, err)
			}
			defer archive.Close()

			hash, err := hashFile(archive)
			if err != nil {
				return fmt.Errorf("failed to hash archive %s: %w", name, err)
			}

			releaseIndexes[name] = indexFile{
				Hash:     hash,
				SelfLink: fmt.Sprintf("https://storage.googleapis.com/openshift-kubebuilder-tools/%s", name),
			}
		}
	}

	index.Releases[version] = releaseIndexes

	indexRaw, err := yaml.Marshal(index)
	if err != nil {
		return fmt.Errorf("failed to marshal index file: %w", err)
	}

	if err := os.WriteFile(indexFileName, indexRaw, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

func hashFile(f *os.File) (string, error) {
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
