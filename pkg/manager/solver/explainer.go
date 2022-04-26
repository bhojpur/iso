package solver

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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	types "github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/crillab/gophersat/bf"
	"github.com/crillab/gophersat/explain"
	"github.com/pkg/errors"
)

type Explainer struct{}

func decodeDimacs(vars map[string]string, dimacs string) (string, error) {
	res := ""
	sc := bufio.NewScanner(bytes.NewBufferString(dimacs))
	lines := strings.Split(dimacs, "\n")
	linenum := 1
SCAN:
	for sc.Scan() {

		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "p":
			continue SCAN
		default:
			for i := 0; i < len(fields)-1; i++ {
				v := fields[i]
				negative := false
				if strings.HasPrefix(fields[i], "-") {
					v = strings.TrimLeft(fields[i], "-")
					negative = true
				}
				variable := vars[v]
				if negative {
					res += fmt.Sprintf("!(%s)", variable)
				} else {
					res += variable
				}

				if i != len(fields)-2 {
					res += fmt.Sprintf(" or ")
				}
			}
			if linenum != len(lines)-1 {
				res += fmt.Sprintf(" and \n")
			}
		}
		linenum++
	}
	if err := sc.Err(); err != nil {
		return res, fmt.Errorf("could not parse problem: %v", err)
	}
	return res, nil
}

func parseVars(r io.Reader) (map[string]string, error) {
	sc := bufio.NewScanner(r)
	res := map[string]string{}
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "c":
			data := strings.Split(fields[1], "=")
			res[data[1]] = data[0]

		default:
			continue

		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("could not parse problem: %v", err)
	}
	return res, nil
}

// Solve tries to find the MUS (minimum unsat) formula from the original problem.
// it returns an error with the decoded dimacs
func (*Explainer) Solve(f bf.Formula, s types.PackageSolver) (types.PackagesAssertions, error) {
	buf := bytes.NewBufferString("")
	if err := bf.Dimacs(f, buf); err != nil {
		return nil, errors.Wrap(err, "cannot extract dimacs from formula")
	}

	copy := *buf

	pb, err := explain.ParseCNF(&copy)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse problem")
	}
	pb2, err := pb.MUS()
	if err != nil {
		return nil, errors.Wrap(err, "could not extract subset")
	}

	variables, err := parseVars(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse variables")
	}

	res, err := decodeDimacs(variables, pb2.CNF())
	if err != nil {
		return nil, errors.Wrap(err, "could not parse dimacs")
	}

	return nil, fmt.Errorf("could not satisfy the constraints: \n%s", res)
}
