<h1>OCM Controller API reference v1alpha1</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#mpas.ocm.software%2fv1alpha1">mpas.ocm.software/v1alpha1</a>
</li>
</ul>
<h2 id="mpas.ocm.software/v1alpha1">mpas.ocm.software/v1alpha1</h2>
<p>Package v1alpha1 contains API Schema definitions for the mpas v1alpha1 API group</p>
Resource Types:
<ul class="simple"></ul>
<h3 id="mpas.ocm.software/v1alpha1.CommitTemplate">CommitTemplate
</h3>
<p>
(<em>Appears on:</em>
<a href="#mpas.ocm.software/v1alpha1.RepositorySpec">RepositorySpec</a>)
</p>
<p>CommitTemplate defines the commit template to use when automated commits are made.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>email</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>name</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="mpas.ocm.software/v1alpha1.Credentials">Credentials
</h3>
<p>
(<em>Appears on:</em>
<a href="#mpas.ocm.software/v1alpha1.RepositorySpec">RepositorySpec</a>)
</p>
<p>Credentials contains ways of authenticating the creation of a repository.</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretRef</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="mpas.ocm.software/v1alpha1.ExistingRepositoryPolicy">ExistingRepositoryPolicy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#mpas.ocm.software/v1alpha1.RepositorySpec">RepositorySpec</a>)
</p>
<p>ExistingRepositoryPolicy defines what to do in case a requested repository already exists.</p>
<h3 id="mpas.ocm.software/v1alpha1.Repository">Repository
</h3>
<p>Repository is the Schema for the repositories API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.RepositorySpec">
RepositorySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>provider</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>owner</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>credentials</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.Credentials">
Credentials
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>defaultBranch</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>visibility</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>isOrganization</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>domain</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Domain specifies an optional domain address to be used instead of the defaults like github.com.
Must NOT contain the scheme.</p>
</td>
</tr>
<tr>
<td>
<code>insecure</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Insecure should be defined if <code>domain</code> is not HTTPS.</p>
</td>
</tr>
<tr>
<td>
<code>maintainers</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>existingRepositoryPolicy</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.ExistingRepositoryPolicy">
ExistingRepositoryPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>commitTemplate</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.CommitTemplate">
CommitTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.RepositoryStatus">
RepositoryStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="mpas.ocm.software/v1alpha1.RepositorySpec">RepositorySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#mpas.ocm.software/v1alpha1.Repository">Repository</a>)
</p>
<p>RepositorySpec defines the desired state of Repository</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>provider</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>owner</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>credentials</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.Credentials">
Credentials
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>defaultBranch</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>interval</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Duration">
Kubernetes meta/v1.Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>visibility</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>isOrganization</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>domain</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Domain specifies an optional domain address to be used instead of the defaults like github.com.
Must NOT contain the scheme.</p>
</td>
</tr>
<tr>
<td>
<code>insecure</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Insecure should be defined if <code>domain</code> is not HTTPS.</p>
</td>
</tr>
<tr>
<td>
<code>maintainers</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>existingRepositoryPolicy</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.ExistingRepositoryPolicy">
ExistingRepositoryPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>commitTemplate</code><br>
<em>
<a href="#mpas.ocm.software/v1alpha1.CommitTemplate">
CommitTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="mpas.ocm.software/v1alpha1.RepositoryStatus">RepositoryStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#mpas.ocm.software/v1alpha1.Repository">Repository</a>)
</p>
<p>RepositoryStatus defines the observed state of Repository</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>observedGeneration</code><br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>ObservedGeneration is the last reconciled generation.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Condition">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
