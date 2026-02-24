// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package conv

import (
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"

	"mlib.com/mrun/strfmt"
)

func TestDateTimeValue(t *testing.T) {
	assert.Equal(t, strfmt.DateTime{}, DateTimeValue(nil))
	time := strfmt.DateTime(time.Now())
	assert.Equal(t, time, DateTimeValue(&time))
}
