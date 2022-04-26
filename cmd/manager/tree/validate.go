package cmd_tree

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
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	helpers "github.com/bhojpur/iso/cmd/manager/helpers"
	"github.com/bhojpur/iso/cmd/manager/util"

	pkg "github.com/bhojpur/iso/pkg/manager/database"
	"github.com/bhojpur/iso/pkg/manager/solver"
	tree "github.com/bhojpur/iso/pkg/manager/tree"

	"github.com/spf13/cobra"
)

type ValidateOpts struct {
	WithSolver    bool
	OnlyRuntime   bool
	OnlyBuildtime bool
	RegExcludes   []*regexp.Regexp
	RegMatches    []*regexp.Regexp
	Excludes      []string
	Matches       []string

	// Runtime validate stuff
	RuntimeCacheDeps *pkg.InMemoryDatabase
	RuntimeReciper   *tree.InstallerRecipe

	// Buildtime validate stuff
	BuildtimeCacheDeps *pkg.InMemoryDatabase
	BuildtimeReciper   *tree.CompilerRecipe

	Mutex      sync.Mutex
	BrokenPkgs int
	BrokenDeps int

	Errors []error
}

func (o *ValidateOpts) IncrBrokenPkgs() {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	o.BrokenPkgs++
}

func (o *ValidateOpts) IncrBrokenDeps() {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	o.BrokenDeps++
}

func (o *ValidateOpts) AddError(err error) {
	o.Mutex.Lock()
	defer o.Mutex.Unlock()
	o.Errors = append(o.Errors, err)
}

func validatePackage(p *types.Package, checkType string, opts *ValidateOpts, reciper tree.Builder, cacheDeps *pkg.InMemoryDatabase) error {
	var errstr string
	var ans error

	var depSolver types.PackageSolver

	if opts.WithSolver {
		emptyInstallationDb := pkg.NewInMemoryDatabase(false)
		depSolver = solver.NewSolver(types.SolverOptions{Type: types.SolverSingleCoreSimple}, pkg.NewInMemoryDatabase(false),
			reciper.GetDatabase(),
			emptyInstallationDb)
	}

	found, err := reciper.GetDatabase().FindPackages(
		&types.Package{
			Name:     p.GetName(),
			Category: p.GetCategory(),
			Version:  ">=0",
		},
	)

	if err != nil || len(found) < 1 {
		if err != nil {
			errstr = err.Error()
		} else {
			errstr = "No packages"
		}
		util.DefaultContext.Error(fmt.Sprintf("[%9s] %s/%s-%s: Broken. No versions could be found by database %s",
			checkType,
			p.GetCategory(), p.GetName(), p.GetVersion(),
			errstr,
		))

		opts.IncrBrokenDeps()

		return errors.New(
			fmt.Sprintf("[%9s] %s/%s-%s: Broken. No versions could be found by database %s",
				checkType,
				p.GetCategory(), p.GetName(), p.GetVersion(),
				errstr,
			))
	}

	// Ensure that we use the right package from right recipier for deps
	pReciper, err := reciper.GetDatabase().FindPackage(
		&types.Package{
			Name:     p.GetName(),
			Category: p.GetCategory(),
			Version:  p.GetVersion(),
		},
	)
	if err != nil {
		errstr = fmt.Sprintf("[%9s] %s/%s-%s: Error on retrieve package - %s.",
			checkType,
			p.GetCategory(), p.GetName(), p.GetVersion(),
			err.Error(),
		)
		util.DefaultContext.Error(errstr)

		return errors.New(errstr)
	}
	p = pReciper

	pkgstr := fmt.Sprintf("%s/%s-%s", p.GetCategory(), p.GetName(),
		p.GetVersion())

	validpkg := true

	if len(opts.Matches) > 0 {
		matched := false
		for _, rgx := range opts.RegMatches {
			if rgx.MatchString(pkgstr) {
				matched = true
				break
			}
		}

		if !matched {
			return nil
		}
	}

	if len(opts.Excludes) > 0 {
		excluded := false
		for _, rgx := range opts.RegExcludes {
			if rgx.MatchString(pkgstr) {
				excluded = true
				break
			}
		}

		if excluded {
			return nil
		}
	}

	util.DefaultContext.Info(fmt.Sprintf("[%9s] Checking package ", checkType)+
		fmt.Sprintf("%s/%s-%s", p.GetCategory(), p.GetName(), p.GetVersion()),
		"with", len(p.GetRequires()), "dependencies and", len(p.GetConflicts()), "conflicts.")

	all := p.GetRequires()
	all = append(all, p.GetConflicts()...)
	for idx, r := range all {

		var deps types.Packages
		var err error
		if r.IsSelector() {
			deps, err = reciper.GetDatabase().FindPackages(
				&types.Package{
					Name:     r.GetName(),
					Category: r.GetCategory(),
					Version:  r.GetVersion(),
				},
			)
		} else {
			deps = append(deps, r)
		}

		if err != nil || len(deps) < 1 {
			if err != nil {
				errstr = err.Error()
			} else {
				errstr = "No packages"
			}
			util.DefaultContext.Error(fmt.Sprintf("[%9s] %s/%s-%s: Broken Dep %s/%s-%s - %s",
				checkType,
				p.GetCategory(), p.GetName(), p.GetVersion(),
				r.GetCategory(), r.GetName(), r.GetVersion(),
				errstr,
			))

			opts.IncrBrokenDeps()

			ans = errors.New(
				fmt.Sprintf("[%9s] %s/%s-%s: Broken Dep %s/%s-%s - %s",
					checkType,
					p.GetCategory(), p.GetName(), p.GetVersion(),
					r.GetCategory(), r.GetName(), r.GetVersion(),
					errstr))

			validpkg = false

		} else {

			util.DefaultContext.Debug(fmt.Sprintf("[%9s] Find packages for dep", checkType),
				fmt.Sprintf("%s/%s-%s", r.GetCategory(), r.GetName(), r.GetVersion()))

			if opts.WithSolver {

				util.DefaultContext.Info(fmt.Sprintf("[%9s]  :soap: [%2d/%2d] %s/%s-%s: %s/%s-%s",
					checkType,
					idx+1, len(all),
					p.GetCategory(), p.GetName(), p.GetVersion(),
					r.GetCategory(), r.GetName(), r.GetVersion(),
				))

				// Check if the solver is already been done for the deep
				_, err := cacheDeps.Get(r.HashFingerprint(""))
				if err == nil {
					util.DefaultContext.Debug(fmt.Sprintf("[%9s]  :direct_hit: Cache Hit for dep", checkType),
						fmt.Sprintf("%s/%s-%s", r.GetCategory(), r.GetName(), r.GetVersion()))
					continue
				}

				util.DefaultContext.Spinner()
				solution, err := depSolver.Install(types.Packages{r})
				ass := solution.SearchByName(r.GetPackageName())
				util.DefaultContext.SpinnerStop()
				if err == nil {
					if ass == nil {

						ans = errors.New(
							fmt.Sprintf("[%9s] %s/%s-%s: solution doesn't retrieve package %s/%s-%s.",
								checkType,
								p.GetCategory(), p.GetName(), p.GetVersion(),
								r.GetCategory(), r.GetName(), r.GetVersion(),
							))

						if util.DefaultContext.Config.General.Debug {
							for idx, pa := range solution {
								fmt.Println(fmt.Sprintf("[%9s] %s/%s-%s: solution %d: %s",
									checkType,
									p.GetCategory(), p.GetName(), p.GetVersion(), idx,
									pa.Package.GetPackageName()))
							}
						}

						util.DefaultContext.Error(ans.Error())
						opts.IncrBrokenDeps()
						validpkg = false
					} else {
						_, err = solution.Order(reciper.GetDatabase(), ass.Package.GetFingerPrint())
					}
				}

				if err != nil {

					util.DefaultContext.Error(fmt.Sprintf("[%9s] %s/%s-%s: solver broken for dep %s/%s-%s - %s",
						checkType,
						p.GetCategory(), p.GetName(), p.GetVersion(),
						r.GetCategory(), r.GetName(), r.GetVersion(),
						err.Error(),
					))

					ans = errors.New(
						fmt.Sprintf("[%9s] %s/%s-%s: solver broken for Dep %s/%s-%s - %s",
							checkType,
							p.GetCategory(), p.GetName(), p.GetVersion(),
							r.GetCategory(), r.GetName(), r.GetVersion(),
							err.Error()))

					opts.IncrBrokenDeps()
					validpkg = false
				}

				// Register the key
				cacheDeps.Set(r.HashFingerprint(""), "1")

			}
		}

	}

	if !validpkg {
		opts.IncrBrokenPkgs()
	}

	return ans
}

func validateWorker(i int,
	wg *sync.WaitGroup,
	c <-chan *types.Package,
	opts *ValidateOpts) {

	defer wg.Done()

	for p := range c {

		if opts.OnlyBuildtime {
			// Check buildtime compiler/deps
			err := validatePackage(p, "buildtime", opts, opts.BuildtimeReciper, opts.BuildtimeCacheDeps)
			if err != nil {
				opts.AddError(err)
				continue
			}
		} else if opts.OnlyRuntime {

			// Check runtime installer/deps
			err := validatePackage(p, "runtime", opts, opts.RuntimeReciper, opts.RuntimeCacheDeps)
			if err != nil {
				opts.AddError(err)
				continue
			}

		} else {

			// Check runtime installer/deps
			err := validatePackage(p, "runtime", opts, opts.RuntimeReciper, opts.RuntimeCacheDeps)
			if err != nil {
				opts.AddError(err)
				continue
			}

			// Check buildtime compiler/deps
			err = validatePackage(p, "buildtime", opts, opts.BuildtimeReciper, opts.BuildtimeCacheDeps)
			if err != nil {
				opts.AddError(err)
			}

		}

	}
}

func initOpts(opts *ValidateOpts, onlyRuntime, onlyBuildtime, withSolver bool, treePaths []string) {
	var err error

	opts.OnlyBuildtime = onlyBuildtime
	opts.OnlyRuntime = onlyRuntime
	opts.WithSolver = withSolver
	opts.RuntimeReciper = nil
	opts.BuildtimeReciper = nil
	opts.BrokenPkgs = 0
	opts.BrokenDeps = 0

	if onlyBuildtime {
		opts.BuildtimeReciper = (tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))).(*tree.CompilerRecipe)
	} else if onlyRuntime {
		opts.RuntimeReciper = (tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))).(*tree.InstallerRecipe)
	} else {
		opts.BuildtimeReciper = (tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))).(*tree.CompilerRecipe)
		opts.RuntimeReciper = (tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))).(*tree.InstallerRecipe)
	}

	opts.RuntimeCacheDeps = pkg.NewInMemoryDatabase(false).(*pkg.InMemoryDatabase)
	opts.BuildtimeCacheDeps = pkg.NewInMemoryDatabase(false).(*pkg.InMemoryDatabase)

	for _, treePath := range treePaths {
		util.DefaultContext.Info(fmt.Sprintf("Loading :deciduous_tree: %s...", treePath))
		if opts.BuildtimeReciper != nil {
			err = opts.BuildtimeReciper.Load(treePath)
			if err != nil {
				util.DefaultContext.Fatal("Error on load tree ", err)
			}
		}
		if opts.RuntimeReciper != nil {
			err = opts.RuntimeReciper.Load(treePath)
			if err != nil {
				util.DefaultContext.Fatal("Error on load tree ", err)
			}
		}
	}

	opts.RegExcludes, err = helpers.CreateRegexArray(opts.Excludes)
	if err != nil {
		util.DefaultContext.Fatal(err.Error())
	}
	opts.RegMatches, err = helpers.CreateRegexArray(opts.Matches)
	if err != nil {
		util.DefaultContext.Fatal(err.Error())
	}

}

func NewTreeValidateCommand() *cobra.Command {
	var excludes []string
	var matches []string
	var treePaths []string
	var opts ValidateOpts

	var ans = &cobra.Command{
		Use:   "validate [OPTIONS]",
		Short: "Validate a tree or a list of packages",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			onlyRuntime, _ := cmd.Flags().GetBool("only-runtime")
			onlyBuildtime, _ := cmd.Flags().GetBool("only-buildtime")

			if len(treePaths) < 1 {
				util.DefaultContext.Fatal("Mandatory tree param missing.")
			}
			if onlyRuntime && onlyBuildtime {
				util.DefaultContext.Fatal("Both --only-runtime and --only-buildtime options are not possibile.")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var reciper tree.Builder

			concurrency := util.DefaultContext.Config.General.Concurrency

			withSolver, _ := cmd.Flags().GetBool("with-solver")
			onlyRuntime, _ := cmd.Flags().GetBool("only-runtime")
			onlyBuildtime, _ := cmd.Flags().GetBool("only-buildtime")

			opts.Excludes = excludes
			opts.Matches = matches
			initOpts(&opts, onlyRuntime, onlyBuildtime, withSolver, treePaths)

			// We need at least one valid reciper for get list of the packages.
			if onlyBuildtime {
				reciper = opts.BuildtimeReciper
			} else {
				reciper = opts.RuntimeReciper
			}

			all := make(chan *types.Package)

			var wg = new(sync.WaitGroup)

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go validateWorker(i, wg, all, &opts)
			}
			for _, p := range reciper.GetDatabase().World() {
				all <- p
			}
			close(all)

			// Wait separately and once done close the channel
			go func() {
				wg.Wait()
			}()

			stringerrs := []string{}
			for _, e := range opts.Errors {
				stringerrs = append(stringerrs, e.Error())
			}
			sort.Strings(stringerrs)
			for _, e := range stringerrs {
				fmt.Println(e)
			}

			// fmt.Println("Broken packages:", brokenPkgs, "(", brokenDeps, "deps ).")
			if len(stringerrs) != 0 {
				util.DefaultContext.Error(fmt.Sprintf("Found %d broken packages and %d broken deps.",
					opts.BrokenPkgs, opts.BrokenDeps))
				util.DefaultContext.Fatal("Errors: " + strconv.Itoa(len(stringerrs)))
			} else {
				util.DefaultContext.Info("All good! :white_check_mark:")
				os.Exit(0)
			}
		},
	}
	path, err := os.Getwd()
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
	ans.Flags().Bool("only-runtime", false, "Check only runtime dependencies.")
	ans.Flags().Bool("only-buildtime", false, "Check only buildtime dependencies.")
	ans.Flags().BoolP("with-solver", "s", false,
		"Enable check of requires also with solver.")
	ans.Flags().StringSliceVarP(&treePaths, "tree", "t", []string{path},
		"Path of the tree to use.")
	ans.Flags().StringSliceVarP(&excludes, "exclude", "e", []string{},
		"Exclude matched packages from analysis. (Use string as regex).")
	ans.Flags().StringSliceVarP(&matches, "matches", "m", []string{},
		"Analyze only matched packages. (Use string as regex).")

	return ans
}
