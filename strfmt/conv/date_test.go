// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package conv

import (
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"

	"mlib.com/mrun/strfmt"
)

func TestDateValue(t *testing.T) {
	assert.Equal(t, strfmt.Date{}, DateValue(nil))
	date := strfmt.Date(time.Now())
	assert.Equal(t, date, DateValue(&date))
}
