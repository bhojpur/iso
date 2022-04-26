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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bhojpur/iso/pkg/manager/api/core/template"
	artifact "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	compression "github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"go.uber.org/multierr"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/compiler"
	"github.com/bhojpur/iso/pkg/manager/installer/client"

	pkg "github.com/bhojpur/iso/pkg/manager/database"
	tree "github.com/bhojpur/iso/pkg/manager/tree"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	REPOSITORY_METAFILE  = "repository.meta.yaml"
	REPOSITORY_SPECFILE  = "repository.yaml"
	TREE_TARBALL         = "tree.tar"
	COMPILERTREE_TARBALL = "compilertree.tar"

	REPOFILE_TREE_KEY          = "tree"
	REPOFILE_COMPILER_TREE_KEY = "compilertree"
	REPOFILE_META_KEY          = "meta"

	DiskRepositoryType   = "disk"
	HttpRepositoryType   = "http"
	DockerRepositoryType = "docker"
)

type BhojpurRepositoryFile struct {
	FileName        string                     `json:"filename"`
	CompressionType compression.Implementation `json:"compressiontype,omitempty"`
	Checksums       artifact.Checksums         `json:"checksums,omitempty"`
}

type BhojpurSystemRepository struct {
	*types.BhojpurRepository

	Index           compiler.ArtifactIndex           `json:"index"`
	BuildTree, Tree tree.Builder                     `json:"-"`
	RepositoryFiles map[string]BhojpurRepositoryFile `json:"repo_files"`
	Backend         compiler.CompilerBackend         `json:"-"`
	PushImages      bool                             `json:"-"`
	ForcePush       bool                             `json:"-"`

	imagePrefix, snapshotID string
}

type BhojpurSystemRepositoryMetadata struct {
	Index []*artifact.PackageArtifact `json:"index,omitempty"`
}

type BhojpurSearchModeType int

const (
	SLabel      = iota
	SRegexPkg   = iota
	SRegexLabel = iota
	FileSearch  = iota
)

type BhojpurSearchOpts struct {
	Mode BhojpurSearchModeType
}

func NewBhojpurSystemRepositoryMetadata(file string, removeFile bool) (*BhojpurSystemRepositoryMetadata, error) {
	ans := &BhojpurSystemRepositoryMetadata{}
	err := ans.ReadFile(file, removeFile)
	if err != nil {
		return nil, err
	}
	return ans, nil
}

// SystemRepositories returns the repositories from the local configuration file
// it filters the available repositories returning the ones that are enabled
func SystemRepositories(t types.BhojpurRepositories) Repositories {
	repos := Repositories{}
	for _, repo := range t.Enabled() {
		r := NewSystemRepository(repo)
		repos = append(repos, r)
	}
	return repos
}

type BuildTreeResult struct {
	Repositories Repositories
	TemplatesDir map[*BhojpurSystemRepository][]string
	RepoDir      map[*BhojpurSystemRepository]string
}

// LoadBuildTree loads to the tree the compilation specs from the system repositories
func LoadBuildTree(t tree.Builder, db types.PackageDatabase, ctx types.Context) (BuildTreeResult, error) {
	var reserr error
	res := &BuildTreeResult{
		TemplatesDir: make(map[*BhojpurSystemRepository][]string),
		RepoDir:      make(map[*BhojpurSystemRepository]string),
	}

	repos := SystemRepositories(ctx.GetConfig().SystemRepositories)
	for _, r := range repos {
		repodir, err := ctx.TempDir(r.Name)
		if err != nil {
			reserr = multierr.Append(reserr, err)
		}
		if err := r.SyncBuildMetadata(ctx, repodir); err != nil {
			reserr = multierr.Append(reserr, err)
		}

		generalRecipe := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))
		if err := generalRecipe.Load(filepath.Join(repodir, "tree")); err != nil {
			reserr = multierr.Append(reserr, err)
		}
		if err := generalRecipe.GetDatabase().Clone(t.GetDatabase()); err != nil {
			reserr = multierr.Append(reserr, err)
		}

		r.SetTree(generalRecipe)
		res.TemplatesDir[r] = template.FindPossibleTemplatesDir(repodir)
		res.RepoDir[r] = ctx.GetConfig().System.GetRepoDatabaseDirPath(r.GetName())
		ctx.Debugf("Loaded repository '%s' with template dir '%s' and repository dir '%s'", r.Name, res.TemplatesDir[r], res.RepoDir[r])
	}

	res.Repositories = repos

	repos.SyncDatabase(db)

	return *res, reserr
}

func (m *BhojpurSystemRepositoryMetadata) WriteFile(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (m *BhojpurSystemRepositoryMetadata) ReadFile(file string, removeFile bool) error {
	if file == "" {
		return errors.New("Invalid path for repository metadata")
	}

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if removeFile {
		defer os.Remove(file)
	}

	err = yaml.Unmarshal(dat, m)
	if err != nil {
		return err
	}

	return nil
}

func (m *BhojpurSystemRepositoryMetadata) ToArtifactIndex() (ans compiler.ArtifactIndex) {
	for _, a := range m.Index {
		ans = append(ans, a)
	}
	return
}

func NewDefaultTreeRepositoryFile() BhojpurRepositoryFile {
	return BhojpurRepositoryFile{
		FileName:        TREE_TARBALL,
		CompressionType: compression.GZip,
	}
}

func NewDefaultCompilerTreeRepositoryFile() BhojpurRepositoryFile {
	return BhojpurRepositoryFile{
		FileName:        COMPILERTREE_TARBALL,
		CompressionType: compression.GZip,
	}
}

func NewDefaultMetaRepositoryFile() BhojpurRepositoryFile {
	return BhojpurRepositoryFile{
		FileName:        REPOSITORY_METAFILE + ".tar",
		CompressionType: compression.None,
	}
}

// SetFileName sets the name of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this set the filename that the client will pull
func (f *BhojpurRepositoryFile) SetFileName(n string) {
	f.FileName = n
}

// GetFileName returns the name of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this gets the filename that the client will pull
func (f *BhojpurRepositoryFile) GetFileName() string {
	return f.FileName
}

// SetCompressionType sets the compression type of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this sets the compression type that the client will use to uncompress the artifact
func (f *BhojpurRepositoryFile) SetCompressionType(c compression.Implementation) {
	f.CompressionType = c
}

// GetCompressionType gets the compression type of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this gets the compression type that the client will use to uncompress the artifact
func (f *BhojpurRepositoryFile) GetCompressionType() compression.Implementation {
	return f.CompressionType
}

// SetChecksums sets the checksum of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this sets the checksums that the client will use to verify the artifact
func (f *BhojpurRepositoryFile) SetChecksums(c artifact.Checksums) {
	f.Checksums = c
}

// GetChecksums gets the checksum of the repository file.
// Each repository can ship arbitrary file that will be downloaded by the client
// in case of need, this gets the checksums that the client will use to verify the artifact
func (f *BhojpurRepositoryFile) GetChecksums() artifact.Checksums {
	return f.Checksums
}

// GenerateRepository generates a new repository from the given argument.
// If the repository is of the docker type, it will also push the package images.
// In case the repository is local, it will build the package Index
func GenerateRepository(p ...RepositoryOption) (*BhojpurSystemRepository, error) {
	c := RepositoryConfig{}
	c.Apply(p...)

	btr := tree.NewCompilerRecipe(pkg.NewInMemoryDatabase(false))
	runtimeTree := pkg.NewInMemoryDatabase(false)

	tempTree := pkg.NewInMemoryDatabase(false)
	temptr := tree.NewInstallerRecipe(tempTree)

	for _, treeDir := range c.Tree {
		if err := temptr.Load(treeDir); err != nil {
			return nil, err
		}
		if err := btr.Load(treeDir); err != nil {
			return nil, err
		}
	}

	// 2: if fromRepo, build a new tree like the compiler is doing and use it to source the above specs,
	// instead of local tree

	repodb := pkg.NewInMemoryDatabase(false)
	generalRecipe := tree.NewCompilerRecipe(repodb)

	if c.FromRepository {
		if _, err := LoadBuildTree(generalRecipe, repodb, c.context); err != nil {
			c.context.Warning("errors while loading trees from repositories", err.Error())
		}
		if err := repodb.Clone(tempTree); err != nil {
			c.context.Warning("errors while cloning trees from repositories", err.Error())
		}
	}

	// Pick only atoms in db which have a real metadata for runtime db (tr)
	for _, p := range tempTree.World() {
		if _, err := os.Stat(filepath.Join(c.Src, p.GetMetadataFilePath())); err == nil {
			runtimeTree.CreatePackage(p)
		}
	}

	// Load packages from metadata files if not present already.
	var ff = func(currentpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only those which are metadata
		if !strings.HasSuffix(info.Name(), types.PackageMetaSuffix) {
			return nil
		}

		dat, err := ioutil.ReadFile(currentpath)
		if err != nil {
			return nil
		}

		art, err := artifact.NewPackageArtifactFromYaml(dat)
		if err != nil {
			return nil
		}
		if _, err := runtimeTree.FindPackage(art.CompileSpec.Package); err != nil && art.CompileSpec.Package.Name != "" {
			c.context.Debug("Adding", art.CompileSpec.Package.HumanReadableString(), "from metadata file", currentpath)
			if art.Runtime != nil && art.Runtime.Name != "" {
				runtimeTree.CreatePackage(art.Runtime)
			} else {
				// We don't have runtime at this point. So we import the package as is
				r := []*types.Package{}
				p := art.CompileSpec.Package.Clone()
				p.Requires(r)
				p.SetProvides(r)
				p.Conflicts(r)
				runtimeTree.CreatePackage(p)
			}
		}

		return nil
	}

	if c.FromMetadata {
		// Best effort
		filepath.Walk(c.Src, ff)
	}

	repo := &BhojpurSystemRepository{
		BhojpurRepository: types.NewBhojpurRepository(c.Name, c.Type, c.Description, c.Urls, c.Priority, true, false),
		Tree:              tree.NewInstallerRecipe(runtimeTree),
		BuildTree:         btr,
		RepositoryFiles:   map[string]BhojpurRepositoryFile{},
		PushImages:        c.PushImages,
		ForcePush:         c.Force,
		Backend:           c.CompilerBackend,
		imagePrefix:       c.ImagePrefix,
	}

	if err := repo.initialize(c.context, c.Src); err != nil {
		return nil, errors.Wrap(err, "while building repository artifact index")
	}

	return repo, nil
}

func NewSystemRepository(repo types.BhojpurRepository) *BhojpurSystemRepository {
	return &BhojpurSystemRepository{
		BhojpurRepository: &repo,
		RepositoryFiles:   map[string]BhojpurRepositoryFile{},
	}
}

func NewBhojpurSystemRepositoryFromYaml(data []byte, db types.PackageDatabase) (*BhojpurSystemRepository, error) {
	var p *BhojpurSystemRepository
	err := yaml.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}

	p.Tree = tree.NewInstallerRecipe(db)

	return p, err
}

func (r *BhojpurSystemRepository) SetPriority(n int) {
	r.BhojpurRepository.Priority = n
}

func (r *BhojpurSystemRepository) initialize(ctx types.Context, src string) error {
	generator, err := r.getGenerator(ctx, r.snapshotID)
	if err != nil {
		return errors.Wrap(err, "while constructing repository generator")
	}
	art, err := generator.Initialize(src, r.Tree.GetDatabase())
	if err != nil {
		return errors.Wrap(err, "while initializing repository generator")
	}
	// update the Bhojpur ISO repository index
	r.Index = art
	return nil
}

// FileSearch search a pattern among the artifacts in a repository
func (r *BhojpurSystemRepository) FileSearch(pattern string) (types.Packages, error) {
	var matches types.Packages
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return matches, err
	}
ARTIFACT:
	for _, a := range r.GetIndex() {
		for _, f := range a.Files {
			if reg.MatchString(f) {
				matches = append(matches, a.CompileSpec.GetPackage())
				continue ARTIFACT
			}
		}
	}
	return matches, nil
}

func (r *BhojpurSystemRepository) GetName() string {
	return r.BhojpurRepository.Name
}
func (r *BhojpurSystemRepository) GetDescription() string {
	return r.BhojpurRepository.Description
}

func (r *BhojpurSystemRepository) GetAuthentication() map[string]string {
	return r.BhojpurRepository.Authentication
}

func (r *BhojpurSystemRepository) GetType() string {
	return r.BhojpurRepository.Type
}

func (r *BhojpurSystemRepository) SetType(p string) {
	r.BhojpurRepository.Type = p
}

// Sets snapshot ID
func (r *BhojpurSystemRepository) SetSnapshotID(i string) {
	r.snapshotID = i
}

func (r *BhojpurSystemRepository) GetVerify() bool {
	return r.BhojpurRepository.Verify
}

func (r *BhojpurSystemRepository) SetVerify(p bool) {
	r.BhojpurRepository.Verify = p
}

func (r *BhojpurSystemRepository) GetBackend() compiler.CompilerBackend {
	return r.Backend
}
func (r *BhojpurSystemRepository) SetBackend(b compiler.CompilerBackend) {
	r.Backend = b
}

func (r *BhojpurSystemRepository) SetName(p string) {
	r.BhojpurRepository.Name = p
}

func (r *BhojpurSystemRepository) AddUrl(p string) {
	r.BhojpurRepository.Urls = append(r.BhojpurRepository.Urls, p)
}
func (r *BhojpurSystemRepository) GetUrls() []string {
	return r.BhojpurRepository.Urls
}
func (r *BhojpurSystemRepository) SetUrls(urls []string) {
	r.BhojpurRepository.Urls = urls
}
func (r *BhojpurSystemRepository) GetPriority() int {
	return r.BhojpurRepository.Priority
}
func (r *BhojpurSystemRepository) GetTreePath() string {
	return r.TreePath
}
func (r *BhojpurSystemRepository) SetTreePath(p string) {
	r.TreePath = p
}
func (r *BhojpurSystemRepository) GetMetaPath() string {
	return r.MetaPath
}
func (r *BhojpurSystemRepository) SetMetaPath(p string) {
	r.MetaPath = p
}
func (r *BhojpurSystemRepository) SetTree(b tree.Builder) {
	r.Tree = b
}
func (r *BhojpurSystemRepository) GetIndex() compiler.ArtifactIndex {
	return r.Index
}
func (r *BhojpurSystemRepository) SetIndex(i compiler.ArtifactIndex) {
	r.Index = i
}
func (r *BhojpurSystemRepository) GetTree() tree.Builder {
	return r.Tree
}
func (r *BhojpurSystemRepository) GetRevision() int {
	return r.BhojpurRepository.Revision
}
func (r *BhojpurSystemRepository) GetLastUpdate() string {
	return r.BhojpurRepository.LastUpdate
}
func (r *BhojpurSystemRepository) SetLastUpdate(u string) {
	r.BhojpurRepository.LastUpdate = u
}
func (r *BhojpurSystemRepository) IncrementRevision() {
	r.BhojpurRepository.Revision++
}
func (r *BhojpurSystemRepository) SetAuthentication(auth map[string]string) {
	r.BhojpurRepository.Authentication = auth
}

// BumpRevision bumps the internal repository revision by reading the current one from repospec
func (r *BhojpurSystemRepository) BumpRevision(repospec string, resetRevision bool) error {
	if resetRevision {
		r.Revision = 0
	} else {
		if _, err := os.Stat(repospec); !os.IsNotExist(err) {
			// Read existing file for retrieve revision
			spec, err := r.ReadSpecFile(repospec)
			if err != nil {
				return err
			}
			r.Revision = spec.GetRevision()
		}
	}
	r.Revision++
	return nil
}

// AddMetadata adds the repository serialized content into the metadata key of the repository
// It writes the serialized content to repospec, and writes the repository.meta.yaml file into dst
func (r *BhojpurSystemRepository) AddMetadata(ctx types.Context, repospec, dst string) (*artifact.PackageArtifact, error) {
	// Create Metadata struct and serialized repository
	meta, serialized := r.Serialize()

	// Create metadata file and repository file
	metaTmpDir, err := ctx.TempDir("metadata")
	defer os.RemoveAll(metaTmpDir) // clean up
	if err != nil {
		return nil, errors.Wrap(err, "Error met while creating tempdir for metadata")
	}

	repoMetaSpec := filepath.Join(metaTmpDir, REPOSITORY_METAFILE)

	// Create repository.meta.yaml file
	err = meta.WriteFile(repoMetaSpec)
	if err != nil {
		return nil, errors.Wrap(err, "failed writing "+repoMetaSpec)
	}

	a, err := r.AddRepositoryFile(metaTmpDir, REPOFILE_META_KEY, dst, NewDefaultMetaRepositoryFile())
	if err != nil {
		return a, errors.Wrap(err, "Error met while adding archive to repository")
	}

	data, err := yaml.Marshal(serialized)
	if err != nil {
		return a, err
	}
	err = ioutil.WriteFile(repospec, data, os.ModePerm)
	if err != nil {
		return a, err
	}
	return a, nil
}

// AddTree adds a tree.Builder with the given key to the repository.
// It will generate an artifact which will be then embedded in the repository manifest
// It returns the generated artifacts and an error
func (r *BhojpurSystemRepository) AddTree(ctx types.Context, t tree.Builder, dst, key string, defaults BhojpurRepositoryFile) (*artifact.PackageArtifact, error) {
	// Create tree and repository file
	archive, err := ctx.TempDir("archive")
	if err != nil {
		return nil, errors.Wrap(err, "Error met while creating tempdir for archive")
	}
	defer os.RemoveAll(archive) // clean up

	if err := t.Save(archive); err != nil {
		return nil, errors.Wrap(err, "Error met while saving the tree")
	}

	a, err := r.AddRepositoryFile(archive, key, dst, defaults)
	if err != nil {
		return nil, errors.Wrap(err, "Error met while adding archive to repository")
	}
	return a, nil
}

// Snapshot creates a copy of the current Bhojpur ISO repository index into dst.
// The copy will be prefixed with "id".
// This allows the clients to refer to old versions of the repository by using the reference_id
func (r *BhojpurSystemRepository) Snapshot(id, dst string) (artifacts []*artifact.PackageArtifact, snapshotIndex string, err error) {

	var snapshotFmt string = "%s-%s"

	repospec := filepath.Join(dst, REPOSITORY_SPECFILE)
	snapshotIndex = filepath.Join(dst, fmt.Sprintf(snapshotFmt, id, REPOSITORY_SPECFILE))

	err = fileHelper.CopyFile(repospec, filepath.Join(dst, snapshotIndex))
	if err != nil {
		err = errors.Wrap(err, "while copying repo spec")
		return
	}

	b, err := ioutil.ReadFile(repospec)
	if err != nil {
		return
	}

	newRepoIndex := &BhojpurSystemRepository{}
	err = yaml.Unmarshal(b, newRepoIndex)
	if err != nil {
		return
	}

	for _, key := range []string{REPOFILE_META_KEY, REPOFILE_TREE_KEY, REPOFILE_COMPILER_TREE_KEY} {
		var isoFile BhojpurRepositoryFile
		isoFile, err = r.GetRepositoryFile(key)
		if err != nil {
			return
		}
		newMetaFile := fmt.Sprintf(snapshotFmt, id, isoFile.FileName)
		err = fileHelper.CopyFile(filepath.Join(dst, isoFile.FileName), filepath.Join(dst, newMetaFile))
		if err != nil {
			return
		}

		m := &isoFile
		m.FileName = newMetaFile
		newRepoIndex.RepositoryFiles[key] = *m
		artifacts = append(artifacts, artifact.NewPackageArtifact(filepath.Join(dst, newMetaFile)))
	}

	_, serialized := newRepoIndex.Serialize()

	data, err := yaml.Marshal(serialized)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(snapshotIndex, data, os.ModePerm)

	return
}

// AddRepositoryFile adds a path to a key in the repository manifest.
// The path will be compressed, and a default File has to be passed in case there is no entry into
// the repository manifest
func (r *BhojpurSystemRepository) AddRepositoryFile(src, fileKey, repositoryRoot string, defaults BhojpurRepositoryFile) (*artifact.PackageArtifact, error) {
	treeFile, err := r.GetRepositoryFile(fileKey)
	if err != nil {
		treeFile = defaults
		//	r.SetRepositoryFile(fileKey, treeFile)
	}

	a := artifact.NewPackageArtifact(filepath.Join(repositoryRoot, treeFile.GetFileName()))
	a.CompressionType = treeFile.GetCompressionType()
	err = a.Compress(src, 1)
	if err != nil {
		return a, errors.Wrap(err, "Error met while creating package archive")
	}

	err = a.Hash()
	if err != nil {
		return a, errors.Wrap(err, "Failed generating checksums for tree")
	}
	// Update the tree name with the name created by compression selected.
	treeFile.SetChecksums(a.Checksums)
	treeFile.SetFileName(path.Base(a.Path))

	r.SetRepositoryFile(fileKey, treeFile)

	return a, nil
}

func (r *BhojpurSystemRepository) GetRepositoryFile(name string) (BhojpurRepositoryFile, error) {
	ans, ok := r.RepositoryFiles[name]
	if ok {
		return ans, nil
	}
	return ans, errors.New("Repository file " + name + " not found!")
}
func (r *BhojpurSystemRepository) SetRepositoryFile(name string, f BhojpurRepositoryFile) {
	r.RepositoryFiles[name] = f
}

func (r *BhojpurSystemRepository) ReadSpecFile(file string) (*BhojpurSystemRepository, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading file "+file)
	}
	var repo *BhojpurSystemRepository
	repo, err = NewBhojpurSystemRepositoryFromYaml(dat, pkg.NewInMemoryDatabase(false))
	if err != nil {
		return nil, errors.Wrap(err, "Error reading repository from file "+file)
	}

	// Check if mandatory key are present
	_, err = repo.GetRepositoryFile(REPOFILE_TREE_KEY)
	if err != nil {
		return nil, errors.New("Invalid repository without the " + REPOFILE_TREE_KEY + " key file.")
	}
	_, err = repo.GetRepositoryFile(REPOFILE_META_KEY)
	if err != nil {
		return nil, errors.New("Invalid repository without the " + REPOFILE_META_KEY + " key file.")
	}

	return repo, err
}

type RepositoryGenerator interface {
	Generate(*BhojpurSystemRepository, string, bool) error
	Initialize(string, types.PackageDatabase) ([]*artifact.PackageArtifact, error)
}

func (r *BhojpurSystemRepository) getGenerator(ctx types.Context, snapshotID string) (RepositoryGenerator, error) {
	if snapshotID == "" {
		snapshotID = time.Now().Format("20060102150405")
	}

	var rg RepositoryGenerator
	switch r.GetType() {
	case DiskRepositoryType, HttpRepositoryType:
		rg = &localRepositoryGenerator{context: ctx, snapshotID: snapshotID}
	case DockerRepositoryType:
		rg = &dockerRepositoryGenerator{
			b:           r.Backend,
			imagePrefix: r.imagePrefix,
			imagePush:   r.PushImages,
			force:       r.ForcePush,
			context:     ctx,
			snapshotID:  snapshotID,
		}
	default:
		return nil, errors.New("invalid repository type")
	}
	return rg, nil
}

// Write writes the repository metadata to the supplied destination
func (r *BhojpurSystemRepository) Write(ctx types.Context, dst string, resetRevision, force bool) error {
	rg, err := r.getGenerator(ctx, r.snapshotID)
	if err != nil {
		return err
	}

	return rg.Generate(r, dst, resetRevision)
}

func (r *BhojpurSystemRepository) Client(ctx types.Context) Client {
	switch r.GetType() {
	case DiskRepositoryType:
		return client.NewLocalClient(client.RepoData{Urls: r.GetUrls()}, ctx)
	case HttpRepositoryType:
		return client.NewHttpClient(
			client.RepoData{
				Urls:           r.GetUrls(),
				Authentication: r.GetAuthentication(),
			}, ctx)

	case DockerRepositoryType:
		return client.NewDockerClient(
			client.RepoData{
				Urls:           r.GetUrls(),
				Authentication: r.GetAuthentication(),
				Verify:         r.Verify,
			}, ctx)
	}
	return nil
}

func (r *BhojpurSystemRepository) SearchArtefact(p *types.Package) (*artifact.PackageArtifact, error) {
	for _, a := range r.GetIndex() {
		if a.CompileSpec.GetPackage().Matches(p) {
			return a, nil
		}
	}

	return nil, errors.New("Not found")
}

func (r *BhojpurSystemRepository) getRepoFile(c Client, key string) (*artifact.PackageArtifact, error) {

	treeFile, err := r.GetRepositoryFile(key)
	if err != nil {
		return nil, errors.Wrapf(err, "key %s not present in the repository", key)
	}

	// Get Tree
	downloadedTreeFile, err := c.DownloadFile(treeFile.GetFileName())
	if err != nil {
		return nil, errors.Wrap(err, "While downloading "+treeFile.GetFileName())
	}
	//defer os.Remove(downloadedTreeFile)

	treeFileArtifact := artifact.NewPackageArtifact(downloadedTreeFile)
	treeFileArtifact.Checksums = treeFile.GetChecksums()
	treeFileArtifact.CompressionType = treeFile.GetCompressionType()

	err = treeFileArtifact.Verify()
	if err != nil {
		return nil, errors.Wrap(err, "file integrity check failure")
	}

	return treeFileArtifact, nil

}

func (r *BhojpurSystemRepository) SyncBuildMetadata(ctx types.Context, path string) error {

	repo, err := r.Sync(ctx, false)
	if err != nil {
		return errors.Wrap(err, "while syncronizing repository")
	}

	c := repo.Client(ctx)
	if c == nil {
		return errors.New("no client could be generated from repository")
	}

	a, err := repo.getRepoFile(c, REPOFILE_COMPILER_TREE_KEY)
	if err != nil {
		return fmt.Errorf("failed while getting: %s", REPOFILE_COMPILER_TREE_KEY)
	}

	defer os.RemoveAll(a.Path)

	if !fileHelper.Exists(filepath.Join(path, "tree")) {
		os.MkdirAll(filepath.Join(path, "tree"), 0600)
	}

	if err := a.Unpack(ctx, filepath.Join(path, "tree"), false); err != nil {
		return errors.Wrapf(err, "while unpacking: %s", REPOFILE_COMPILER_TREE_KEY)
	}

	for _, ai := range repo.GetTree().GetDatabase().World() {
		// Retrieve remote repository.yaml for retrieve revision and date
		file, err := c.DownloadFile(ai.GetMetadataFilePath())
		if err != nil {
			return errors.Wrapf(err, "while downloading metadata for %s", ai.HumanReadableString())
		}
		if err := fileHelper.Move(file, filepath.Join(path, ai.GetMetadataFilePath())); err != nil {
			return err
		}
	}

	return nil
}

func (r *BhojpurSystemRepository) referenceID() string {
	repositoryReferenceID := REPOSITORY_SPECFILE
	if r.ReferenceID != "" {
		repositoryReferenceID = r.ReferenceID
	}
	return repositoryReferenceID
}

func (r *BhojpurSystemRepository) Sync(ctx types.Context, force bool) (*BhojpurSystemRepository, error) {
	var repoUpdated bool = false
	var treefs, metafs string

	repobasedir := ctx.GetConfig().System.GetRepoDatabaseDirPath(r.GetName())

	toTimeSync := false
	dat, err := ioutil.ReadFile(filepath.Join(repobasedir, "SYNCTIME"))
	if err == nil {
		parsed, _ := time.Parse(time.RFC3339, string(dat))
		if time.Now().After(parsed.Add(24 * time.Hour)) {
			toTimeSync = true
			ctx.Debug(r.Name, "is old, refresh is suggested")
		}
	} else {
		toTimeSync = true
	}

	ctx.Debug("Sync of the repository", r.Name, "in progress...")

	c := r.Client(ctx)
	if c == nil {
		return nil, errors.New("no client could be generated from repository")
	}

	repositoryReferenceID := r.referenceID()

	var downloadedRepoMeta *BhojpurSystemRepository
	var file string
	repoFile := filepath.Join(repobasedir, repositoryReferenceID)

	_, repoExistsErr := os.Stat(repoFile)
	if toTimeSync || force || os.IsNotExist(repoExistsErr) {
		// Retrieve remote repository.yaml for retrieve revision and date
		file, err = c.DownloadFile(repositoryReferenceID)
		if err != nil {
			return nil, errors.Wrap(err, "while downloading "+repositoryReferenceID)
		}
		downloadedRepoMeta, err = r.ReadSpecFile(file)
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(file)
		defer func() {
			now := time.Now().Format(time.RFC3339)
			ioutil.WriteFile(filepath.Join(repobasedir, "SYNCTIME"), []byte(now), os.ModePerm)
		}()
	} else {
		downloadedRepoMeta, err = r.ReadSpecFile(repoFile)
		if err != nil {
			return nil, err
		}
		repoUpdated = true
	}

	if r.Cached {
		if !force {
			localRepo, _ := r.ReadSpecFile(filepath.Join(repobasedir, repositoryReferenceID))
			if localRepo != nil {
				if localRepo.GetRevision() == downloadedRepoMeta.GetRevision() &&
					localRepo.GetLastUpdate() == downloadedRepoMeta.GetLastUpdate() {
					repoUpdated = true
				}
			}
		}
		if r.GetTreePath() == "" {
			treefs = filepath.Join(repobasedir, "treefs")
		} else {
			treefs = r.GetTreePath()
		}
		if r.GetMetaPath() == "" {
			metafs = filepath.Join(repobasedir, "metafs")
		} else {
			metafs = r.GetMetaPath()
		}

	} else {
		treefs, err = ctx.TempDir("treefs")
		if err != nil {
			return nil, errors.Wrap(err, "Error met while creating tempdir for rootfs")
		}
		metafs, err = ctx.TempDir("metafs")
		if err != nil {
			return nil, errors.Wrap(err, "Error met whilte creating tempdir for metafs")
		}
	}

	// treeFile and metaFile must be present, they aren't optional
	if !repoUpdated {

		treeFileArtifact, err := downloadedRepoMeta.getRepoFile(c, REPOFILE_TREE_KEY)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching '%s'", REPOFILE_TREE_KEY)
		}
		defer os.Remove(treeFileArtifact.Path)

		ctx.Debug("Tree tarball for the repository " + r.GetName() + " downloaded correctly.")

		metaFileArtifact, err := downloadedRepoMeta.getRepoFile(c, REPOFILE_META_KEY)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching '%s'", REPOFILE_META_KEY)
		}
		defer os.Remove(metaFileArtifact.Path)

		ctx.Debug("Metadata tarball for the repository " + r.GetName() + " downloaded correctly.")

		if r.Cached {
			// Copy updated repository.yaml file to repo dir now that the tree is synced.
			err = fileHelper.CopyFile(file, filepath.Join(repobasedir, repositoryReferenceID))
			if err != nil {
				return nil, errors.Wrap(err, "Error on update "+repositoryReferenceID)
			}
			// Remove previous tree
			os.RemoveAll(treefs)
			// Remove previous meta dir
			os.RemoveAll(metafs)
		}
		ctx.Debug("Decompress tree of the repository " + r.Name + "...")

		if _, err := os.Lstat(treefs); os.IsNotExist(err) {
			os.MkdirAll(treefs, 0600)
		}

		err = treeFileArtifact.Unpack(ctx, treefs, false)
		if err != nil {
			return nil, errors.Wrap(err, "Error met while unpacking tree")
		}

		// FIXME: It seems that tar with only one file doesn't create destination
		//       directory. I create directory directly for now.
		os.MkdirAll(metafs, os.ModePerm)
		err = metaFileArtifact.Unpack(ctx, metafs, false)
		if err != nil {
			return nil, errors.Wrap(err, "Error met while unpacking metadata")
		}

		tsec, _ := strconv.ParseInt(downloadedRepoMeta.GetLastUpdate(), 10, 64)

		ctx.Info(
			fmt.Sprintf(":house: Repository %s revision: %d (%s)",
				downloadedRepoMeta.GetName(),
				downloadedRepoMeta.GetRevision(),
				time.Unix(tsec, 0).String()))

	}

	meta, err := NewBhojpurSystemRepositoryMetadata(
		filepath.Join(metafs, REPOSITORY_METAFILE), false,
	)
	if err != nil {
		return nil, errors.Wrap(err, "While processing "+REPOSITORY_METAFILE)
	}
	downloadedRepoMeta.SetIndex(meta.ToArtifactIndex())

	reciper := tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))
	err = reciper.Load(treefs)
	if err != nil {
		return nil, errors.Wrap(err, "Error met while unpacking rootfs")
	}

	downloadedRepoMeta.SetTree(reciper)
	downloadedRepoMeta.SetTreePath(treefs)

	// Copy the local available data to the one which was synced
	// e.g. locally we can override the type (disk), or priority
	// while remotely it could be advertized differently
	r.fill(downloadedRepoMeta)

	if !repoUpdated {
		ctx.Info(
			fmt.Sprintf(":information_source: Repository: %s Priority: %d Type: %s",
				downloadedRepoMeta.GetName(),
				downloadedRepoMeta.GetPriority(),
				downloadedRepoMeta.GetType()))
	}
	return downloadedRepoMeta, nil
}

func (r *BhojpurSystemRepository) fill(r2 *BhojpurSystemRepository) {
	r2.SetUrls(r.GetUrls())
	r2.SetAuthentication(r.GetAuthentication())
	r2.SetType(r.GetType())
	r2.SetPriority(r.GetPriority())
	r2.SetName(r.GetName())
	r2.SetVerify(r.GetVerify())
}

func (r *BhojpurSystemRepository) Serialize() (*BhojpurSystemRepositoryMetadata, BhojpurSystemRepository) {

	serialized := *r
	serialized.Authentication = nil

	serialized.Index = compiler.ArtifactIndex{}

	meta := &BhojpurSystemRepositoryMetadata{
		Index: []*artifact.PackageArtifact{},
	}
	for _, a := range r.Index {
		cp := *a
		copy := &cp
		copy.Path = filepath.Base(copy.Path)
		meta.Index = append(meta.Index, copy)
	}

	return meta, serialized
}

func (r Repositories) Len() int      { return len(r) }
func (r Repositories) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Repositories) Less(i, j int) bool {
	return r[i].GetPriority() < r[j].GetPriority()
}

func (r Repositories) World() types.Packages {
	cache := map[string]*types.Package{}
	world := types.Packages{}

	// Get Uniques. Walk in reverse so the definitions of most prio-repo overwrites lower ones
	// In this way, when we will walk again later the deps sorting them by most higher prio we have better chance of success.
	for i := len(r) - 1; i >= 0; i-- {
		for _, p := range r[i].GetTree().GetDatabase().World() {
			cache[p.GetFingerPrint()] = p
		}
	}

	for _, v := range cache {
		world = append(world, v)
	}

	return world
}

func (r Repositories) SyncDatabase(d types.PackageDatabase) {
	cache := map[string]bool{}

	// Get Uniques. Walk in reverse so the definitions of most prio-repo overwrites lower ones
	// In this way, when we will walk again later the deps sorting them by most higher prio we have better chance of success.
	for i := len(r) - 1; i >= 0; i-- {
		for _, p := range r[i].GetTree().GetDatabase().World() {
			if _, ok := cache[p.GetFingerPrint()]; !ok {
				cache[p.GetFingerPrint()] = true
				d.CreatePackage(p)
			}
		}
	}
}

type PackageMatch struct {
	Repo     *BhojpurSystemRepository
	Artifact *artifact.PackageArtifact
	Package  *types.Package
}

func (re Repositories) PackageMatches(p types.Packages) []PackageMatch {
	// TODO: Better heuristic. here we pick the first repo that contains the atom, sorted by priority but
	// we should do a permutations and get the best match, and in case there are more solutions the user should be able to pick
	sort.Sort(re)

	var matches []PackageMatch
PACKAGE:
	for _, pack := range p {
		for _, r := range re {
			c, err := r.GetTree().GetDatabase().FindPackage(pack)
			if err == nil {
				matches = append(matches, PackageMatch{Package: c, Repo: r})
				continue PACKAGE
			}
		}
	}

	return matches

}

func (re Repositories) ResolveSelectors(p types.Packages) types.Packages {
	// If a selector is given, get the best from each repo
	sort.Sort(re) // respect prio
	var matches types.Packages
PACKAGE:
	for _, pack := range p {
	REPOSITORY:
		for _, r := range re {
			if pack.IsSelector() {
				c, err := r.GetTree().GetDatabase().FindPackageCandidate(pack)
				// If FindPackageCandidate returns the same package, it means it couldn't find one.
				// Skip this repository and keep looking.
				if err != nil { //c.String() == pack.String() {
					continue REPOSITORY
				}
				matches = append(matches, c)
				continue PACKAGE
			} else {
				// If it's not a selector, just append it
				matches = append(matches, pack)
			}
		}
	}

	return matches

}

func (re Repositories) SearchPackages(p string, t BhojpurSearchModeType) []PackageMatch {
	sort.Sort(re)
	var matches []PackageMatch
	var err error

	for _, r := range re {
		var repoMatches types.Packages

		switch t {
		case SRegexPkg:
			repoMatches, err = r.GetTree().GetDatabase().FindPackageMatch(p)
		case SLabel:
			repoMatches, err = r.GetTree().GetDatabase().FindPackageLabel(p)
		case SRegexLabel:
			repoMatches, err = r.GetTree().GetDatabase().FindPackageLabelMatch(p)
		case FileSearch:
			repoMatches, err = r.FileSearch(p)
		}

		if err == nil && len(repoMatches) > 0 {
			for _, pack := range repoMatches {
				a, _ := r.SearchArtefact(pack)
				matches = append(matches, PackageMatch{Package: pack, Repo: r, Artifact: a})
			}
		}
	}

	return matches
}

func (re Repositories) SearchLabelMatch(s string) []PackageMatch {
	return re.SearchPackages(s, SRegexLabel)
}

func (re Repositories) SearchLabel(s string) []PackageMatch {
	return re.SearchPackages(s, SLabel)
}

func (re Repositories) Search(s string) []PackageMatch {
	return re.SearchPackages(s, SRegexPkg)
}
