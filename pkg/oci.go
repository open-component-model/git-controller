// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package pkg

import "context"

// OCIClient defines the needed capabilities of a client that can interact with an OCI repository.
type OCIClient interface {
	Pull(ctx context.Context, url, outDir string) (string, error)
}
