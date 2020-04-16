// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&bitCacherTestSuite{})

type bitCacherTestSuite struct {
	r *mux.Router
}

func (g *bitCacherTestSuite) SetUpTest(c *C) {
	g.r = makeRouter()
}

func (g *bitCacherTestSuite) TestIntegration(c *C) {
	buf := `{"action": "create", "name": "some-name"}`

	req, err := http.NewRequest("GET", "/api/1/get/some-name", nil)
	respRec := httptest.NewRecorder()
	g.r.ServeHTTP(respRec, req)
	c.Assert(respRec.Code, Equals, 404)

	req, err = http.NewRequest("POST", "/api/1/change", bytes.NewBufferString(buf))
	c.Assert(err, IsNil)

	respRec = httptest.NewRecorder()
	g.r.ServeHTTP(respRec, req)
	c.Assert(respRec.Code, Equals, 201)

	req, err = http.NewRequest("GET", "/api/1/get/some-name", nil)
	respRec = httptest.NewRecorder()
	g.r.ServeHTTP(respRec, req)
	c.Assert(respRec.Code, Equals, 200)
}
