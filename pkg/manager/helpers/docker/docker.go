package docker

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"context"
	"encoding/hex"
	"os"

	bhojpurimages "github.com/bhojpur/iso/pkg/manager/api/core/image"
	bhojpurtypes "github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/containerd/containerd/images"

	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	"github.com/bhojpur/iso/pkg/manager/api/core/bus"
	"github.com/docker/cli/cli/trust"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/theupdateframework/notary/tuf/data"
)

// See also https://github.com/docker/cli/blob/88c6089300a82d3373892adf6845a4fed1a4ba8d/cli/command/image/trust.go#L171

func verifyImage(image string, authConfig *types.AuthConfig) (string, error) {
	ref, err := reference.ParseAnyReference(image)
	if err != nil {
		return "", errors.Wrapf(err, "invalid reference %s", image)
	}

	// only check if image ref doesn't contain hashes
	if _, ok := ref.(reference.Digested); !ok {
		namedRef, ok := ref.(reference.Named)
		if !ok {
			return "", errors.New("failed to resolve image digest using content trust: reference is not named")
		}
		namedRef = reference.TagNameOnly(namedRef)
		taggedRef, ok := namedRef.(reference.NamedTagged)
		if !ok {
			return "", errors.New("failed to resolve image digest using content trust: reference is not tagged")
		}

		resolvedImage, err := trustedResolveDigest(context.Background(), taggedRef, authConfig, "bhojpur")
		if err != nil {
			return "", errors.Wrap(err, "failed to resolve image digest using content trust")
		}
		resolvedFamiliar := reference.FamiliarString(resolvedImage)
		return resolvedFamiliar, nil
	}

	return "", nil
}

func trustedResolveDigest(ctx context.Context, ref reference.NamedTagged, authConfig *types.AuthConfig, useragent string) (reference.Canonical, error) {
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return nil, err
	}

	notaryRepo, err := trust.GetNotaryRepository(os.Stdin, os.Stdout, useragent, repoInfo, authConfig, "pull")
	if err != nil {
		return nil, errors.Wrap(err, "error establishing connection to trust repository")
	}

	t, err := notaryRepo.GetTargetByName(ref.Tag(), trust.ReleasesRole, data.CanonicalTargetsRole)
	if err != nil {
		return nil, trust.NotaryError(repoInfo.Name.Name(), err)
	}
	// Only get the tag if it's in the top level targets role or the releases delegation role
	// ignore it if it's in any other delegation roles
	if t.Role != trust.ReleasesRole && t.Role != data.CanonicalTargetsRole {
		return nil, trust.NotaryError(repoInfo.Name.Name(), errors.Errorf("No trust data for %s", reference.FamiliarString(ref)))
	}

	h, ok := t.Hashes["sha256"]
	if !ok {
		return nil, errors.New("no valid hash, expecting sha256")
	}

	dgst := digest.NewDigestFromHex("sha256", hex.EncodeToString(h))

	// Allow returning canonical reference with tag
	return reference.WithDigest(ref, dgst)
}

type staticAuth struct {
	auth *types.AuthConfig
}

func (s staticAuth) Authorization() (*authn.AuthConfig, error) {
	if s.auth == nil {
		return nil, nil
	}
	return &authn.AuthConfig{
		Username:      s.auth.Username,
		Password:      s.auth.Password,
		Auth:          s.auth.Auth,
		IdentityToken: s.auth.IdentityToken,
		RegistryToken: s.auth.RegistryToken,
	}, nil
}

// UnpackEventData is the data structure to pass for the bus events
type UnpackEventData struct {
	Image string
	Dest  string
}

// DownloadAndExtractDockerImage extracts a container image natively. It supports privileged/unprivileged mode
func DownloadAndExtractDockerImage(ctx bhojpurtypes.Context, image, dest string, auth *types.AuthConfig, verify bool) (*images.Image, error) {
	if verify {
		img, err := verifyImage(image, auth)
		if err != nil {
			return nil, errors.Wrapf(err, "failed verifying image")
		}
		image = img
	}

	if !fileHelper.Exists(dest) {
		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "cannot create destination directory")
		}
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, remote.WithAuth(staticAuth{auth}))
	if err != nil {
		return nil, err
	}

	m, err := img.Manifest()
	if err != nil {
		return nil, err
	}

	mt, err := img.MediaType()
	if err != nil {
		return nil, err
	}

	d, err := img.Digest()
	if err != nil {
		return nil, err
	}

	bus.Manager.Publish(bus.EventImagePreUnPack, UnpackEventData{Image: image, Dest: dest})

	var c int64
	c, _, err = bhojpurimages.ExtractTo(
		ctx,
		img,
		dest,
		nil,
	)
	if err != nil {
		return nil, err
	}

	bus.Manager.Publish(bus.EventImagePostUnPack, UnpackEventData{Image: image, Dest: dest})

	return &images.Image{
		Name:   image,
		Labels: m.Annotations,
		Target: specs.Descriptor{
			MediaType: string(mt),
			Digest:    digest.Digest(d.String()),
			Size:      c,
		},
	}, nil
}

func ExtractDockerImage(ctx bhojpurtypes.Context, local, dest string) (*images.Image, error) {
	if !fileHelper.Exists(dest) {
		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "cannot create destination directory")
		}
	}

	ref, err := name.ParseReference(local)
	if err != nil {
		return nil, err
	}

	img, err := daemon.Image(ref)
	if err != nil {
		return nil, err
	}

	m, err := img.Manifest()
	if err != nil {
		return nil, err
	}

	mt, err := img.MediaType()
	if err != nil {
		return nil, err
	}

	d, err := img.Digest()
	if err != nil {
		return nil, err
	}

	var c int64
	c, _, err = bhojpurimages.ExtractTo(
		ctx,
		img,
		dest,
		nil,
	)

	if err != nil {
		return nil, err
	}

	bus.Manager.Publish(bus.EventImagePostUnPack, UnpackEventData{Image: local, Dest: dest})

	return &images.Image{
		Name:   local,
		Labels: m.Annotations,
		Target: specs.Descriptor{
			MediaType: string(mt),
			Digest:    digest.Digest(d.String()),
			Size:      c,
		},
	}, nil
}