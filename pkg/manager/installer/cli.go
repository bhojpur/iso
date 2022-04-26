package installer

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
	"fmt"
	"sort"
	"strings"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/pterm/pterm"
)

func packsToList(p types.Packages) string {
	var packs []string

	for _, pp := range p {
		packs = append(packs, pp.HumanReadableString())
	}

	sort.Strings(packs)
	return strings.Join(packs, " ")
}

func printList(p types.Packages) {
	fmt.Println()
	d := pterm.TableData{{"Program Name", "Version", "License"}}
	for _, m := range p {
		d = append(d, []string{
			fmt.Sprintf("%s/%s", m.GetCategory(), m.GetName()),
			pterm.LightGreen(m.GetVersion()), m.GetLicense()})
	}
	pterm.DefaultTable.WithHasHeader().WithData(d).Render()
	fmt.Println()
}

func printUpgradeList(install, uninstall types.Packages) {
	fmt.Println()

	d := pterm.TableData{{"Old version", "New version", "License"}}
	for _, m := range uninstall {
		if p, err := install.Find(m.GetPackageName()); err == nil {
			d = append(d, []string{
				pterm.LightRed(m.HumanReadableString()),
				pterm.LightGreen(p.HumanReadableString()), m.GetLicense()})
		} else {
			d = append(d, []string{
				pterm.LightRed(m.HumanReadableString()), ""})
		}
	}
	for _, m := range install {
		if _, err := uninstall.Find(m.GetPackageName()); err != nil {
			d = append(d, []string{"",
				pterm.LightGreen(m.HumanReadableString()), m.GetLicense()})
		}
	}
	pterm.DefaultTable.WithHasHeader().WithData(d).Render()
	fmt.Println()

}

func printMatchUpgrade(artefacts map[string]ArtifactMatch, uninstall types.Packages) {
	p := types.Packages{}

	for _, a := range artefacts {
		p = append(p, a.Package)
	}

	printUpgradeList(p, uninstall)
}

func printMatches(artefacts map[string]ArtifactMatch) {
	fmt.Println()
	d := pterm.TableData{{"Program Name", "Version", "License", "Repository"}}
	for _, m := range artefacts {
		d = append(d, []string{
			fmt.Sprintf("%s/%s", m.Package.GetCategory(), m.Package.GetName()),
			pterm.LightGreen(m.Package.GetVersion()), m.Package.GetLicense(), m.Repository.GetName()})
	}
	pterm.DefaultTable.WithHasHeader().WithData(d).Render()
	fmt.Println()
}
