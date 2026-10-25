package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrActionNotAllowed = errors.New("ACTION_NOT_ALLOWED")

type ChecksumRunner struct{}

func (ChecksumRunner) Execute(_ context.Context, action string) (ArtifactEvidence, error) {
	if action != "local.checksum" {
		return ArtifactEvidence{}, ErrActionNotAllowed
	}
	digest := sha256.Sum256([]byte(publicFixture))
	return ArtifactEvidence{
		SHA256:     hex.EncodeToString(digest[:]),
		ByteCount:  int64(len(publicFixture)),
		Provenance: "local_public_fixture_v1",
	}, nil
}
