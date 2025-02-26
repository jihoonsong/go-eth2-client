// Copyright © 2024 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package electra

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/attestantio/go-eth2-client/codecs"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/deneb"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

// beaconBlockBodyJSON is the spec representation of the struct.
type beaconBlockBodyJSON struct {
	RANDAOReveal          phase0.BLSSignature                   `json:"randao_reveal"`
	ETH1Data              *phase0.ETH1Data                      `json:"eth1_data"`
	Graffiti              string                                `json:"graffiti"`
	ProposerSlashings     []*phase0.ProposerSlashing            `json:"proposer_slashings"`
	AttesterSlashings     []*AttesterSlashing                   `json:"attester_slashings"`
	Attestations          []*Attestation                        `json:"attestations"`
	Deposits              []*phase0.Deposit                     `json:"deposits"`
	VoluntaryExits        []*phase0.SignedVoluntaryExit         `json:"voluntary_exits"`
	SyncAggregate         *altair.SyncAggregate                 `json:"sync_aggregate"`
	ExecutionPayload      *deneb.ExecutionPayload               `json:"execution_payload"`
	BLSToExecutionChanges []*capella.SignedBLSToExecutionChange `json:"bls_to_execution_changes"`
	BlobKZGCommitments    []string                              `json:"blob_kzg_commitments"`
	ExecutionRequests     *ExecutionRequests                    `json:"execution_requests"`
}

// MarshalJSON implements json.Marshaler.
func (b *BeaconBlockBody) MarshalJSON() ([]byte, error) {
	blobKZGCommitments := make([]string, len(b.BlobKZGCommitments))
	for i := range b.BlobKZGCommitments {
		blobKZGCommitments[i] = b.BlobKZGCommitments[i].String()
	}

	return json.Marshal(&beaconBlockBodyJSON{
		RANDAOReveal:          b.RANDAOReveal,
		ETH1Data:              b.ETH1Data,
		Graffiti:              fmt.Sprintf("%#x", b.Graffiti),
		ProposerSlashings:     b.ProposerSlashings,
		AttesterSlashings:     b.AttesterSlashings,
		Attestations:          b.Attestations,
		Deposits:              b.Deposits,
		VoluntaryExits:        b.VoluntaryExits,
		SyncAggregate:         b.SyncAggregate,
		ExecutionPayload:      b.ExecutionPayload,
		BLSToExecutionChanges: b.BLSToExecutionChanges,
		BlobKZGCommitments:    blobKZGCommitments,
		ExecutionRequests:     b.ExecutionRequests,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
//nolint:gocyclo
func (b *BeaconBlockBody) UnmarshalJSON(input []byte) error {
	raw, err := codecs.RawJSON(&beaconBlockBodyJSON{}, input)
	if err != nil {
		return err
	}

	if err := b.RANDAOReveal.UnmarshalJSON(raw["randao_reveal"]); err != nil {
		return errors.Wrap(err, "randao_reveal")
	}

	if err := json.Unmarshal(raw["eth1_data"], &b.ETH1Data); err != nil {
		return errors.Wrap(err, "eth1_data")
	}

	graffiti := raw["graffiti"]
	if !bytes.HasPrefix(graffiti, []byte{'"', '0', 'x'}) {
		return errors.New("graffiti: invalid prefix")
	}
	if !bytes.HasSuffix(graffiti, []byte{'"'}) {
		return errors.New("graffiti: invalid suffix")
	}
	if len(graffiti) != 1+2+32*2+1 {
		return errors.New("graffiti: incorrect length")
	}
	length, err := hex.Decode(b.Graffiti[:], graffiti[3:3+32*2])
	if err != nil {
		return errors.Wrap(err, "graffiti")
	}
	if length != 32 {
		return errors.New("graffiti: incorrect length")
	}

	if err := json.Unmarshal(raw["proposer_slashings"], &b.ProposerSlashings); err != nil {
		return errors.Wrap(err, "proposer_slashings")
	}
	for i := range b.ProposerSlashings {
		if b.ProposerSlashings[i] == nil {
			return fmt.Errorf("proposer slashings entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["attester_slashings"], &b.AttesterSlashings); err != nil {
		return errors.Wrap(err, "attester_slashings")
	}
	for i := range b.AttesterSlashings {
		if b.AttesterSlashings[i] == nil {
			return fmt.Errorf("attester slashings entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["attestations"], &b.Attestations); err != nil {
		return errors.Wrap(err, "attestations")
	}
	for i := range b.Attestations {
		if b.Attestations[i] == nil {
			return fmt.Errorf("attestations entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["deposits"], &b.Deposits); err != nil {
		return errors.Wrap(err, "deposits")
	}
	for i := range b.Deposits {
		if b.Deposits[i] == nil {
			return fmt.Errorf("deposits entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["voluntary_exits"], &b.VoluntaryExits); err != nil {
		return errors.Wrap(err, "voluntary_exits")
	}
	for i := range b.VoluntaryExits {
		if b.VoluntaryExits[i] == nil {
			return fmt.Errorf("voluntary exits entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["sync_aggregate"], &b.SyncAggregate); err != nil {
		return errors.Wrap(err, "sync_aggregate")
	}

	if err := json.Unmarshal(raw["execution_payload"], &b.ExecutionPayload); err != nil {
		return errors.Wrap(err, "execution_payload")
	}

	if err := json.Unmarshal(raw["bls_to_execution_changes"], &b.BLSToExecutionChanges); err != nil {
		return errors.Wrap(err, "bls_to_execution_changes")
	}
	for i := range b.BLSToExecutionChanges {
		if b.BLSToExecutionChanges[i] == nil {
			return fmt.Errorf("bls to execution changes entry %d missing", i)
		}
	}

	if err := json.Unmarshal(raw["blob_kzg_commitments"], &b.BlobKZGCommitments); err != nil {
		return errors.Wrap(err, "blob_kzg_commitments")
	}

	if err := json.Unmarshal(raw["execution_requests"], &b.ExecutionRequests); err != nil {
		return errors.Wrap(err, "execution_requests")
	}

	return nil
}
